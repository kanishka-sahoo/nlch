# Homebrew Formula for nlch
# This file can be used to create a Homebrew tap for easier installation on macOS
#
# To use:
# 1. Create a GitHub repository named "homebrew-nlch"
# 2. Place this file at "Formula/nlch.rb"
# 3. Users can then install with: brew tap kanishka-sahoo/nlch && brew install nlch

class Nlch < Formula
  desc "Natural Language Command Helper - Generate terminal commands from natural language"
  homepage "https://github.com/kanishka-sahoo/nlch"
  version "0.1.0" # This will be updated automatically by the release workflow
  
  # These URLs and checksums will be updated automatically for each release
  if OS.mac?
    if Hardware::CPU.arm?
      url "https://github.com/kanishka-sahoo/nlch/releases/download/v#{version}/nlch-darwin-arm64"
      sha256 "REPLACE_WITH_ACTUAL_ARM64_SHA256"
    else
      url "https://github.com/kanishka-sahoo/nlch/releases/download/v#{version}/nlch-darwin-amd64"
      sha256 "REPLACE_WITH_ACTUAL_AMD64_SHA256"
    end
  end
  
  def install
    if OS.mac?
      if Hardware::CPU.arm?
        bin.install "nlch-darwin-arm64" => "nlch"
      else
        bin.install "nlch-darwin-amd64" => "nlch"
      end
    end
  end
  
  test do
    assert_match "nlch version", shell_output("#{bin}/nlch --version")
  end
end
