class Archon < Formula
  desc "AI Orchestration Harness CLI - Install structured AI workflows per project"
  homepage "https://github.com/AlexCas/archon-ai"
  url "https://github.com/AlexCas/archon-ai/archive/refs/tags/v0.4.0.tar.gz"
  sha256 "f09637a62e2493f9a329173b1d87cfd9b1f0d86b76783e5945f5c7b7ae09176f"
  version "0.4.0"
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
