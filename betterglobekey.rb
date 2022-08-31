class Betterglobekey < Formula
  desc "Make macOS Globe key great again!"
  version "latest"
  homepage "https://github.com/Serpentiel/betterglobekey"
  url "https://github.com/Serpentiel/betterglobekey.git", tag: "latest"
  license "MIT"
  head "https://github.com/Serpentiel/betterglobekey.git", branch: "main"

  depends_on "go"

  def install
    system "go", "build", "-o", "#{bin}/betterglobekey", "./main.go"

    output = Utils.safe_popen_read("#{bin}/betterglobekey", "completion", "bash")
    (bash_completion/"betterglobekey").write output

    output = Utils.safe_popen_read("#{bin}/betterglobekey", "completion", "zsh")
    (zsh_completion/"_betterglobekey").write output

    output = Utils.safe_popen_read("#{bin}/betterglobekey", "completion", "fish")
    (fish_completion/"betterglobekey.fish").write output
  end

  service do
    run "#{bin}/betterglobekey"
    keep_alive true
  end

  test do
    str_default = shell_output("#{bin}/betterglobekey")
    str_help = shell_output("#{bin}/betterglobekey --help")
    assert_equal str_default, str_help

    assert_match "Usage:", str_help
    assert_match "Available Commands:", str_help
  end
end
