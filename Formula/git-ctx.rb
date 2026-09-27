class GitCtx < Formula
  desc "Distributed, offline-first context storage embedded in git"
  homepage "https://github.com/jxucoder/git-context"
  version "0.3.1"
  license "Apache-2.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/jxucoder/git-context/releases/download/v0.3.1/git-ctx-darwin-arm64"
      sha256 "f1386d38d983b7c4c131fb59e1c9c36e8068d7840ebecba64cd9100954533849"
    else
      url "https://github.com/jxucoder/git-context/releases/download/v0.3.1/git-ctx-darwin-amd64"
      sha256 "aa2a9e1545b88d57f925e47bece855f9d44363c0980e5a9e9a07c06c4fa9b3cd"
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

