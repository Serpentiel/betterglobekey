// The control contract is compiled straight from the buf module at the
// repository root, so the companion can never drift from the daemon.
const PROTO_ROOT: &str = "../../proto";
const PROTO_FILE: &str = "betterglobekey/control/v1/control.proto";

fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("cargo:rerun-if-changed={PROTO_ROOT}/{PROTO_FILE}");

    let descriptors = protox::compile([PROTO_FILE], [PROTO_ROOT])?;

    tonic_prost_build::configure()
        // The server is only generated for the integration tests' stub daemon.
        .build_server(true)
        // Messages cross the IPC boundary as JSON shaped like src/shared/types.ts.
        .type_attribute(
            ".betterglobekey.control.v1",
            "#[derive(serde::Serialize, serde::Deserialize)] #[serde(rename_all = \"camelCase\", default)]",
        )
        .compile_fds(descriptors)?;

    tauri_build::build();

    Ok(())
}
