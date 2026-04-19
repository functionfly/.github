# Homebrew formula for the fly CLI.
# To use: brew tap functionfly/tap && brew install fly
# Or copy this file to your tap repo and update version/sha256 on each release.
# Checksums: https://github.com/functionfly/functionfly/releases

class Fly < Formula
  desc "FunctionFly CLI - Go from idea to global API in under 60 seconds"
  homepage "https://functionfly.com"
  version "1.2.0"
  license "Apache-2.0"

  on_macos do
    on_arm do
      url "https://github.com/functionfly/functionfly/releases/download/v#{version}/fly_#{version}_macos_arm64.tar.gz"
      sha256 "576918025bd7b819e899ab4f2d5b9f8be2cbaaaa33c1fc7e90a1ad5490541b23"
    end
    on_intel do
      url "https://github.com/functionfly/functionfly/releases/download/v#{version}/fly_#{version}_macos_x86_64.tar.gz"
      sha256 "d210e46734140136f1cda66d727fae9fdedcab1b85a1c4772b01578413f48ff0"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/functionfly/functionfly/releases/download/v#{version}/fly_#{version}_linux_arm64.tar.gz"
      sha256 "7f8d6a08b89217de43d844f825b615ebf969d4fcfff1f1107459bff25d91cb54"
    end
    on_intel do
      url "https://github.com/functionfly/functionfly/releases/download/v#{version}/fly_#{version}_linux_x86_64.tar.gz"
      sha256 "1a56a4da5b854c9cc935c2e5714a1b690f6e8c2fc253e4c378fdc08ce40fef44"
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
