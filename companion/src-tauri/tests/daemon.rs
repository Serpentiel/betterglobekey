//! Exercises the client end to end against a stub daemon on a real Unix socket.

use std::sync::{Arc, Mutex};

use betterglobekey_companion::daemon::Daemon;
use betterglobekey_companion::daemon::proto::config_service_server::{ConfigService, ConfigServiceServer};
use betterglobekey_companion::daemon::proto::*;
use tokio::net::UnixListener;
use tokio_stream::wrappers::UnixListenerStream;
use tonic::transport::Server;
use tonic::{Request, Response, Status};

#[derive(Default)]
struct Stub {
    applied: Arc<Mutex<Option<Config>>>,
}

#[tonic::async_trait]
impl ConfigService for Stub {
    async fn get_config(&self, _: Request<GetConfigRequest>) -> Result<Response<GetConfigResponse>, Status> {
        // Leave every section but collections unset to exercise complete().
        let config = Config {
            collections: vec![Collection { name: "primary".into(), sources: vec!["com.apple.keylayout.US".into()] }],
            ..Default::default()
        };

        Ok(Response::new(GetConfigResponse { config: Some(config) }))
    }

    async fn apply_config(
        &self,
        request: Request<ApplyConfigRequest>,
    ) -> Result<Response<ApplyConfigResponse>, Status> {
        let config = request.into_inner().config.unwrap_or_default();

        if config.collections.is_empty() {
            return Err(Status::invalid_argument("at least one collection is required"));
        }

        *self.applied.lock().unwrap() = Some(config);

        Ok(Response::new(ApplyConfigResponse {}))
    }

    async fn list_input_sources(
        &self,
        _: Request<ListInputSourcesRequest>,
    ) -> Result<Response<ListInputSourcesResponse>, Status> {
        let sources = vec![InputSource { id: "com.apple.keylayout.US".into(), name: "U.S.".into() }];

        Ok(Response::new(ListInputSourcesResponse { sources }))
    }

    async fn get_version(&self, _: Request<GetVersionRequest>) -> Result<Response<GetVersionResponse>, Status> {
        Ok(Response::new(GetVersionResponse { version: "4.0.1".into(), commit: "abc1234".into() }))
    }

    async fn list_keyboards(
        &self,
        _: Request<ListKeyboardsRequest>,
    ) -> Result<Response<ListKeyboardsResponse>, Status> {
        let keyboard = Keyboard { id: "046d:c33f".into(), name: "Logitech".into(), listed: true, ..Default::default() };

        Ok(Response::new(ListKeyboardsResponse { keyboards: vec![keyboard], system_mappings: vec![] }))
    }

    async fn record_key(&self, request: Request<RecordKeyRequest>) -> Result<Response<RecordKeyResponse>, Status> {
        let keyboard_id = request.into_inner().keyboard_id;

        if keyboard_id.is_empty() {
            return Err(Status::failed_precondition("reading raw keyboard input failed"));
        }

        Ok(Response::new(RecordKeyResponse { key: "insert".into(), key_name: "Insert".into(), keyboard_id }))
    }

    async fn get_status(&self, _: Request<GetStatusRequest>) -> Result<Response<GetStatusResponse>, Status> {
        Ok(Response::new(GetStatusResponse { accessibility_trusted: true, input_monitoring_granted: true }))
    }
}

/// serve starts a stub daemon on a fresh socket and returns a client for it.
fn serve(stub: Stub) -> (Daemon, tempfile::TempDir) {
    let dir = tempfile::tempdir().unwrap();
    let socket = dir.path().join("control.sock");
    let incoming = UnixListenerStream::new(UnixListener::bind(&socket).unwrap());

    tokio::spawn(Server::builder().add_service(ConfigServiceServer::new(stub)).serve_with_incoming(incoming));

    (Daemon::connect(socket), dir)
}

#[tokio::test]
async fn reads_config_sources_and_version() {
    let (daemon, _dir) = serve(Stub::default());

    let config = daemon.get_config().await.unwrap();
    assert_eq!(config.collections[0].name, "primary");
    assert!(
        config.logger.is_some() && config.double_press.is_some() && config.reverse.is_some() && config.hud.is_some()
    );
    // The stub predates the trigger, so the Globe default fills it in.
    let trigger = config.trigger.unwrap();
    assert_eq!((trigger.key.as_str(), trigger.mode.as_str()), ("globe", "fn"));
    assert!(trigger.keyboards.is_empty());

    assert_eq!(daemon.list_keyboards().await.unwrap().keyboards[0].name, "Logitech");

    let recorded = daemon.record_key("046d:c33f".into()).await.unwrap();
    assert_eq!((recorded.key.as_str(), recorded.keyboard_id.as_str()), ("insert", "046d:c33f"));
    assert_eq!(daemon.record_key(String::new()).await.unwrap_err(), "reading raw keyboard input failed");

    assert_eq!(daemon.list_input_sources().await.unwrap()[0].name, "U.S.");
    assert_eq!(daemon.get_version().await.unwrap().version, "4.0.1");
}

#[tokio::test]
async fn applies_config() {
    let stub = Stub::default();
    let applied = stub.applied.clone();
    let (daemon, _dir) = serve(stub);

    let mut config = daemon.get_config().await.unwrap();
    config.hud.as_mut().unwrap().enabled = true;
    daemon.apply_config(config.clone()).await.unwrap();

    assert_eq!(applied.lock().unwrap().as_ref(), Some(&config));
}

#[tokio::test]
async fn surfaces_the_daemons_error_message() {
    let (daemon, _dir) = serve(Stub::default());

    let error = daemon.apply_config(Config::default()).await.unwrap_err();

    assert_eq!(error, "at least one collection is required");
}

#[tokio::test]
async fn names_the_socket_when_the_daemon_is_not_running() {
    let dir = tempfile::tempdir().unwrap();
    let daemon = Daemon::connect(dir.path().join("missing.sock"));

    let error = daemon.get_config().await.unwrap_err();

    assert_eq!(error, format!("cannot connect to {}", daemon.socket().display()));
}
