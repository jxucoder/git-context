class GitCtx < Formula
  desc "Distributed, offline-first context storage embedded in git"
  homepage "https://github.com/jxucoder/git-context"
  version "0.1.0"
  license "Apache-2.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jxucoder/git-context/releases/download/v0.1.0/git-ctx-darwin-arm64"
      sha256 "0b72f1343583b957eec8e3251fbd043faf9b27035f8699efdc9d37f6f81ee42b"
    else
      url "https://github.com/jxucoder/git-context/releases/download/v0.1.0/git-ctx-darwin-amd64"
      sha256 "615d99c1e3fdf105821179de3c20566c427331448e6c0c7ad41c857f76073974"
    end
  end

  def install
    binary_name = Hardware::CPU.arm? ? "git-ctx-darwin-arm64" : "git-ctx-darwin-amd64"
    bin.install binary_name => "git-ctx"
  end

  def post_install
    system "git", "config", "--global", "alias.ctx", "!git-ctx"
  end

  test do
    system "#{bin}/git-ctx", "--version"
  end
end

