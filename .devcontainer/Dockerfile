FROM mcr.microsoft.com/devcontainers/base:debian

# mise owns the toolchain; the versions live in mise.toml, which Renovate
# already tracks. Nothing is pinned twice.
RUN install -dm 755 /etc/apt/keyrings \
 && apt-get update \
 && apt-get install -y --no-install-recommends gpg wget curl ca-certificates \
 && wget -qO - https://mise.jdx.dev/gpg-key.pub | gpg --dearmor > /etc/apt/keyrings/mise-archive-keyring.gpg \
 && echo "deb [signed-by=/etc/apt/keyrings/mise-archive-keyring.gpg arch=$(dpkg --print-architecture)] https://mise.jdx.dev/deb stable main" > /etc/apt/sources.list.d/mise.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends mise \
 && rm -rf /var/lib/apt/lists/*

# Created here so a cache volume mounted over them keeps vscode's ownership.
# The image stays root: envbuilder runs the container that way, and switching
# users here leaves mise's lock files owned by the wrong one. devcontainer.json
# still asks for vscode, which is what an editor attaches as.
RUN mkdir -p /home/vscode/.local/share/mise /home/vscode/.local/state/mise \
 && chown -R vscode:vscode /home/vscode/.local \
 && echo 'eval "$(mise activate bash)"' >> /home/vscode/.bashrc \
 && echo 'eval "$(mise activate zsh)"' >> /home/vscode/.zshrc \
 && echo 'eval "$(mise activate bash)"' >> /root/.bashrc
