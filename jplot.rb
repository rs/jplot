class Jplot < Formula
  desc "iTerm2 expvar/JSON monitoring tool"
  homepage "https://github.com/rs/jplot"
  url "https://github.com/rs/jplot/archive/1.0.0.tar.gz"
  sha256 "f6816198294c67e5d858effbfb744097a9c1520a494a1da6699b2a99422151d1"
  head "https://github.com/rs/jplot.git"

  if Hardware::CPU.is_64_bit?
    url "https://github.com/rs/jplot/releases/download/1.0.0/jplot_1.0.0_darwin_amd64.zip"
    sha256 "13529d71da748903de3e7e034722d30c4ef9cf1329f2dab366547d8afed4ea7e"
  else
    url "https://github.com/rs/jplot/releases/download/1.0.0/jplot_1.0.0_darwin_386.zip"
    sha256 "3581dc8488e45467ff4ec089a1c74216cea7e26e97091247f31f04ee79dd6c22"
  end

  depends_on "go" => :build

  def install
    bin.install "jplot"
  end
end
