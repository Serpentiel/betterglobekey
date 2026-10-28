//! Client for the daemon's control API, served over a Unix domain socket in the
//! user's home directory.

use std::path::{Path, PathBuf};
use std::time::Duration;

use hyper_util::rt::TokioIo;
use tokio::net::UnixStream;
use tonic::transport::{Channel, Endpoint, Uri};
use tonic::{Code, Request, Status};
use tower::service_fn;

pub mod proto {
    tonic::include_proto!("betterglobekey.control.v1");
}

use proto::config_service_client::ConfigServiceClient;
use proto::{
    ApplyConfigRequest, Config, GetConfigRequest, GetVersionRequest, GetVersionResponse, InputSource,
    ListInputSourcesRequest, ListKeyboardsRequest, ListKeyboardsResponse, RecordKeyRequest, RecordKeyResponse,
};

/// Daemon is a client for the daemon's ConfigService. It is cheap to clone.
#[derive(Clone)]
pub struct Daemon {
    client: ConfigServiceClient<Channel>,
    socket: PathBuf,
}

impl Daemon {
    /// connect returns a client for the control socket at `socket`. It dials on
    /// first use and redials after a failure, so the daemon may start after the
    /// companion. It must be called within a Tokio runtime.
    pub fn connect(socket: impl Into<PathBuf>) -> Self {
        let socket = socket.into();
        let path = socket.clone();

        // tonic requires a URI, but every connection dials the socket instead.
        let channel = Endpoint::from_static("http://localhost")
            .connect_timeout(Duration::from_secs(2))
            .connect_with_connector_lazy(service_fn(move |_: Uri| {
                let path = path.clone();

                async move { UnixStream::connect(path).await.map(TokioIo::new) }
            }));

        Self { client: ConfigServiceClient::new(channel), socket }
    }

    /// socket returns the path of the control socket this client dials.
    pub fn socket(&self) -> &Path {
        &self.socket
    }

    pub async fn get_config(&self) -> Result<Config, String> {
        let response = self.client.clone().get_config(GetConfigRequest {}).await.map_err(|s| self.error(s))?;

        Ok(complete(response.into_inner().config.unwrap_or_default()))
    }

    pub async fn apply_config(&self, config: Config) -> Result<(), String> {
        let request = ApplyConfigRequest { config: Some(config) };
        self.client.clone().apply_config(request).await.map_err(|s| self.error(s))?;

        Ok(())
    }

    pub async fn list_input_sources(&self) -> Result<Vec<InputSource>, String> {
        let response =
            self.client.clone().list_input_sources(ListInputSourcesRequest {}).await.map_err(|s| self.error(s))?;

        Ok(response.into_inner().sources)
    }

    pub async fn list_keyboards(&self) -> Result<ListKeyboardsResponse, String> {
        let response = self.client.clone().list_keyboards(ListKeyboardsRequest {}).await.map_err(|s| self.error(s))?;

        Ok(response.into_inner())
    }

    /// record_key returns the next key pressed on the keyboard with
    /// `keyboard_id`, or on any keyboard when it is empty. The daemon reads it
    /// from raw input, so it waits until a key is pressed or RECORD_TIMEOUT.
    pub async fn record_key(&self, keyboard_id: String) -> Result<RecordKeyResponse, String> {
        let mut request = Request::new(RecordKeyRequest { keyboard_id });
        request.set_timeout(RECORD_TIMEOUT);

        let response = self.client.clone().record_key(request).await.map_err(|s| match s.code() {
            Code::DeadlineExceeded => "no key was pressed in time".to_owned(),
            _ => self.error(s),
        })?;

        Ok(response.into_inner())
    }

    pub async fn get_version(&self) -> Result<GetVersionResponse, String> {
        let response = self.client.clone().get_version(GetVersionRequest {}).await.map_err(|s| self.error(s))?;

        Ok(response.into_inner())
    }

    /// error turns a gRPC status into the message the UI shows. A transport
    /// failure means the daemon is not listening, so name the socket instead of
    /// surfacing the connector's error chain.
    fn error(&self, status: Status) -> String {
        match status.code() {
            Code::Unavailable => format!("cannot connect to {}", self.socket.display()),
            _ => status.message().to_owned(),
        }
    }
}

/// RECORD_TIMEOUT is how long record_key waits for a key.
const RECORD_TIMEOUT: Duration = Duration::from_secs(30);

/// default_socket returns the daemon's control socket in the user's home
/// directory.
pub fn default_socket() -> PathBuf {
    std::env::home_dir().unwrap_or_default().join(".betterglobekey.sock")
}

/// complete fills in any section the daemon left unset, so the UI always gets
/// every field (proto3 omits empty messages; the UI types do not).
fn complete(mut config: Config) -> Config {
    config.logger.get_or_insert_default();
    config.double_press.get_or_insert_default();
    config.reverse.get_or_insert_default();
    config.hud.get_or_insert_default();
    // An unset Globe key switches, as the daemon treats it.
    config.trigger.get_or_insert_default().globe.get_or_insert(true);

    config
}

#[cfg(test)]
mod tests {
    use serde_json::json;

    use super::proto::{KeyMapping, Keyboard, KeyboardTrigger, KeyboardTriggers, Trigger};
    use super::*;

    #[test]
    fn complete_fills_unset_sections() {
        let config = complete(Config::default());

        assert_eq!(
            serde_json::to_value(config).unwrap(),
            json!({
                "logger": { "path": "", "level": "", "retentionDays": 0, "retentionFiles": 0 },
                "doublePress": { "enabled": false, "maximumDelay": "" },
                "reverse": { "enabled": false, "modifier": "" },
                "hud": { "enabled": false, "duration": "", "showCollection": false },
                "collections": [],
                "trigger": { "keyboards": {}, "globe": true },
            }),
        );
    }

    #[test]
    fn config_deserializes_from_the_ui_shape() {
        let config: Config = serde_json::from_value(json!({
            "logger": { "path": "/tmp/b.log", "level": "info", "retentionDays": 7, "retentionFiles": 3 },
            "doublePress": { "enabled": true, "maximumDelay": "250ms" },
            "reverse": { "enabled": true, "modifier": "shift" },
            "hud": { "enabled": true, "duration": "900ms", "showCollection": true },
            "collections": [{ "name": "primary", "sources": ["com.apple.keylayout.US"] }],
            "trigger": {
                "globe": false,
                "keyboards": {
                    "046d:c33f": [{ "key": "caps_lock", "mode": "dedicated" }, { "key": "insert", "mode": "pass_through" }],
                    "05ac:0281": [],
                },
            },
        }))
        .unwrap();

        assert_eq!(config.logger.unwrap().retention_days, 7);
        assert_eq!(config.double_press.unwrap().maximum_delay, "250ms");
        assert!(config.hud.unwrap().show_collection);
        assert_eq!(config.collections[0].sources, ["com.apple.keylayout.US"]);

        let trigger = config.trigger.unwrap();
        assert_eq!(trigger.globe, Some(false));
        let logitech = &trigger.keyboards["046d:c33f"].keys;
        assert_eq!(logitech.len(), 2);
        assert_eq!((logitech[1].key.as_str(), logitech[1].mode.as_str()), ("insert", "pass_through"));
        assert!(trigger.keyboards["05ac:0281"].keys.is_empty());
    }

    #[test]
    fn complete_keeps_a_set_trigger() {
        let config = complete(Config {
            trigger: Some(Trigger {
                globe: None,
                keyboards: [(
                    "046d:c33f".to_owned(),
                    KeyboardTriggers {
                        keys: vec![KeyboardTrigger { key: "insert".to_owned(), mode: "pass_through".to_owned() }],
                    },
                )]
                .into(),
            }),
            ..Config::default()
        });

        let value = serde_json::to_value(config).unwrap();
        assert_eq!(value["trigger"]["keyboards"]["046d:c33f"], json!([{ "key": "insert", "mode": "pass_through" }]));
        assert_eq!(value["trigger"]["globe"], json!(true));
    }

    #[test]
    fn keyboards_serialize_to_the_ui_shape() {
        let response = ListKeyboardsResponse {
            keyboards: vec![Keyboard {
                id: "046d:c33f".to_owned(),
                triggers: vec![KeyboardTrigger { key: "caps_lock".to_owned(), mode: "dedicated".to_owned() }],
                listed: true,
                trigger_mappings: vec![KeyMapping {
                    src: 0x7_0000_0039,
                    dst: 0x7_0000_0073,
                    src_name: "Caps Lock".to_owned(),
                    dst_name: "F24".to_owned(),
                }],
                ..Keyboard::default()
            }],
            system_mappings: vec![],
            input_monitoring_granted: false,
        };

        let value = serde_json::to_value(response).unwrap();
        assert_eq!(value["keyboards"][0]["triggers"], json!([{ "key": "caps_lock", "mode": "dedicated" }]));
        assert_eq!(value["keyboards"][0]["listed"], json!(true));
        assert_eq!(value["keyboards"][0]["triggerMappings"][0]["srcName"], json!("Caps Lock"));
        assert_eq!(value["keyboards"][0]["otherMappings"], json!([]));
        assert_eq!(value["systemMappings"], json!([]));
        assert_eq!(value["inputMonitoringGranted"], json!(false));
    }
}
