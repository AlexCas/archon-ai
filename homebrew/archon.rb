class Archon < Formula
  desc "AI Orchestration Harness CLI - Install structured AI workflows per project"
  homepage "https://github.com/AlexCas/archon-ai"
  url "https://github.com/AlexCas/archon-ai/archive/refs/tags/v0.3.0.tar.gz"
  sha256 "e466b8773d7aba332108313267238af3dab96de79bf040ac7c29502620b626b4"
  version "0.3.0"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", 
      "-ldflags", 
      "-X github.com/archon-ai/archon/internal/version.Version=#{version} -X github.com/archon-ai/archon/internal/version.Commit=brew -X github.com/archon-ai/archon/internal/version.Date=#{Time.now.iso8601}",
      "-o", bin/"archon",
      "./cmd/archon"
  end

  test do
    assert_match "archon version #{version}", shell_output("#{bin}/archon version")
  end
end
