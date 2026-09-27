class GitCtx < Formula
  desc "Distributed, offline-first context storage embedded in git"
  homepage "https://github.com/jxucoder/git-context"
  version "0.3.0"
  license "Apache-2.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jxucoder/git-context/releases/download/v0.3.0/git-ctx-darwin-arm64"
      sha256 "5495666ce58740f0fee333ab135960a9858fd1b2aaeeaadbd7bf77bbe4c35b1b"
    else
      url "https://github.com/jxucoder/git-context/releases/download/v0.3.0/git-ctx-darwin-amd64"
      sha256 "9ad6c00ebb17c31b6c12311ff85f954daaafa5eeb9e0a75ea19303f9514c7ef8"
    end
  end

  def install
    binary_name = Hardware::CPU.arm? ? "git-ctx-darwin-arm64" : "git-ctx-darwin-amd64"
    bin.install binary_name => "git-ctx"
  end

  def caveats
    <<~EOS
      To use as `git ctx`, add the alias:
        git config --global alias.ctx '!git-ctx'
    EOS
  end

  test do
    system "#{bin}/git-ctx", "--version"
  end
end

