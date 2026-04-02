# Homebrew formula for the fly CLI.
# To use: brew tap functionfly/tap && brew install fly
# Or copy this file to your tap repo and update version/sha256 on each release.
# Checksums: https://github.com/functionfly/functionfly/releases

class Fly < Formula
  desc "FunctionFly CLI - Go from idea to global API in under 60 seconds"
  homepage "https://functionfly.com"
  version "0.0.0"
  license "Apache-2.0"

  on_macos do
    on_arm do
      url "https://github.com/functionfly/functionfly/releases/download/v#{version}/fly_#{version}_macos_arm64.tar.gz"
      sha256 "placeholder_arm64_macos"
    end
    on_intel do
      url "https://github.com/functionfly/functionfly/releases/download/v#{version}/fly_#{version}_macos_x86_64.tar.gz"
      sha256 "placeholder_x86_64_macos"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/functionfly/functionfly/releases/download/v#{version}/fly_#{version}_linux_arm64.tar.gz"
      sha256 "placeholder_arm64_linux"
    end
    on_intel do
      url "https://github.com/functionfly/functionfly/releases/download/v#{version}/fly_#{version}_linux_x86_64.tar.gz"
      sha256 "placeholder_x86_64_linux"
    end
  end

  def install
    # GoReleaser wrap_in_directory: archive contains fly_VER_OS_ARCH/fly
    dir = Dir.glob("fly_*").find { |d| File.directory?(d) }
    if dir && File.executable?("#{dir}/fly")
      bin.install "#{dir}/fly" => "fly"
    else
      bin.install "fly"
    end
  end

  test do
    assert_match "fly version", shell_output("#{bin}/fly version 2>&1")
  end
end
