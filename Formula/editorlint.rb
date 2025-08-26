class Editorlint < Formula
  desc "A comprehensive tool to validate and fix files according to .editorconfig specifications"
  homepage "https://github.com/cdobbyn/editorlint"
  license "MIT"
  head "https://github.com/cdobbyn/editorlint.git", branch: "main"

  # This will be updated automatically by the release workflow
  url "https://github.com/cdobbyn/editorlint/archive/refs/tags/1.3.3.tar.gz"
  sha256 "fd8219691c2ec7a45a0bc8ec8c51a1a55a4419aaa6fb2a0c9d3f19bc4a740841"
  version "1.3.3"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/editorlint"
  end

  test do
    # Create a test .editorconfig file
    (testpath/".editorconfig").write <<~EOS
      root = true

      [*]
      insert_final_newline = true
      trim_trailing_whitespace = true
    EOS

    # Create a test file with violations
    (testpath/"test.txt").write "test with trailing spaces   "

    # Run editorlint and expect it to find violations
    output = shell_output("#{bin}/editorlint test.txt", 1)
    assert_match "validation failed", output
    assert_match "trim_trailing_whitespace", output

    # Test the fix functionality
    system bin/"editorlint", "--fix", "test.txt"

    # Verify the file was fixed
    fixed_content = File.read(testpath/"test.txt")
    assert_equal "test with trailing spaces\n", fixed_content
  end
end
