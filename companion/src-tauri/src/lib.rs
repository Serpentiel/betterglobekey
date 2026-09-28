//! The betterglobekey companion: a Tauri shell around the React UI, with the
//! daemon's gRPC control API exposed to it as commands.

pub mod daemon;

use tauri::webview::PageLoadEvent;
use tauri::window::Color;
use tauri::{AppHandle, Manager, State, Url, WebviewUrl, WebviewWindowBuilder, WindowEvent};
use tauri_plugin_opener::OpenerExt;

use daemon::Daemon;
use daemon::proto::{Config, GetVersionResponse, InputSource};

const MAIN_WINDOW: &str = "main";

#[tauri::command]
async fn get_config(daemon: State<'_, Daemon>) -> Result<Config, String> {
    daemon.get_config().await
}

#[tauri::command]
async fn apply_config(daemon: State<'_, Daemon>, config: Config) -> Result<(), String> {
    daemon.apply_config(config).await
}

#[tauri::command]
async fn list_input_sources(daemon: State<'_, Daemon>) -> Result<Vec<InputSource>, String> {
    daemon.list_input_sources().await
}

#[tauri::command]
async fn get_version(daemon: State<'_, Daemon>) -> Result<GetVersionResponse, String> {
    daemon.get_version().await
}

/// is_app_url reports whether `url` belongs to the app itself (the bundled UI,
/// or the Vite dev server in debug builds) rather than the outside web.
fn is_app_url(url: &Url) -> bool {
    let bundled = url.scheme() == "tauri" || url.host_str() == Some("tauri.localhost");
    let dev_server = cfg!(debug_assertions) && url.scheme() == "http" && url.host_str() == Some("localhost");

    bundled || dev_server
}

fn create_window(app: &AppHandle) -> tauri::Result<()> {
    let opener = app.clone();

    let builder = WebviewWindowBuilder::new(app, MAIN_WINDOW, WebviewUrl::default())
        .title("betterglobekey")
        .inner_size(820.0, 760.0)
        .min_inner_size(560.0, 520.0)
        .background_color(Color(0, 0, 0, 255))
        // Stay hidden until the first paint so the window never flashes blank.
        .visible(false)
        .on_page_load(|window, payload| {
            if payload.event() == PageLoadEvent::Finished {
                let _ = window.show();
            }
        })
        // Never navigate away from the app; hand anything else to the browser.
        .on_navigation(move |url| {
            if is_app_url(url) {
                return true;
            }

            let _ = opener.opener().open_url(url.as_str(), None::<&str>);

            false
        });

    #[cfg(target_os = "macos")]
    let builder = builder
        .title_bar_style(tauri::TitleBarStyle::Overlay)
        .hidden_title(true)
        .traffic_light_position(tauri::LogicalPosition::new(19.0, 18.0));

    builder.build()?;

    Ok(())
}

/// show_main_window brings the window back from hidden, minimized, or behind
/// other apps.
fn show_main_window(app: &AppHandle) {
    if let Some(window) = app.get_webview_window(MAIN_WINDOW) {
        let _ = window.show();
        let _ = window.unminimize();
        let _ = window.set_focus();
    }
}

pub fn run() {
    let app = tauri::Builder::default()
        // Registered first, so a second launch exits before anything else starts.
        .plugin(tauri_plugin_single_instance::init(|app, _args, _cwd| show_main_window(app)))
        .plugin(tauri_plugin_opener::init())
        .invoke_handler(tauri::generate_handler![get_config, apply_config, list_input_sources, get_version])
        .setup(|app| {
            let daemon = tauri::async_runtime::block_on(async { Daemon::connect(daemon::default_socket()) });
            app.manage(daemon);

            create_window(app.handle())?;

            Ok(())
        })
        // Like any macOS app, closing the window keeps the app running; the Dock
        // icon brings the window back.
        .on_window_event(|window, event| {
            if let WindowEvent::CloseRequested { api, .. } = event
                && cfg!(target_os = "macos")
            {
                api.prevent_close();
                let _ = window.hide();
            }
        })
        .build(tauri::generate_context!())
        .expect("failed to build the companion");

    app.run(|app, event| {
        #[cfg(target_os = "macos")]
        if let tauri::RunEvent::Reopen { .. } = event {
            show_main_window(app);
        }

        #[cfg(not(target_os = "macos"))]
        let _ = (app, event);
    });
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn app_urls_stay_in_the_window() {
        for url in ["tauri://localhost/index.html", "http://tauri.localhost/"] {
            assert!(is_app_url(&Url::parse(url).unwrap()), "{url}");
        }

        assert_eq!(is_app_url(&Url::parse("http://localhost:5173/").unwrap()), cfg!(debug_assertions));
    }

    #[test]
    fn external_urls_leave_the_window() {
        for url in ["https://github.com/Serpentiel/betterglobekey", "http://example.com/", "file:///etc/passwd"] {
            assert!(!is_app_url(&Url::parse(url).unwrap()), "{url}");
        }
    }
}
