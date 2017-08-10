package brew

import "text/template"

var formula = template.Must(template.New("rb").Parse(`
require "language/go"

class {{ .FormulaName }}< Formula
  desc "{{ .Description }}"
  homepage "{{ .Homepage }}"
  url "https://github.com/{{ .Owner }}/{{ .Name }}/archive/{{ .Tag }}.tar.gz"
  sha256 "{{ .Checksum }}"
  head "https://github.com/{{ .Owner }}/{{ .Name }}.git"

  depends_on "glide" => :build
  depends_on "go" => :build

  bottle do
    root_url "{{ .BottleURL }}"
    cellar :any_skip_relocation
{{range .Bottles }}    sha256 "{{ .Checksum }}" => :{{ .Name }}
{{ end }}  end

  go_resource "github.com/jteeuwen/go-bindata" do
    url "https://github.com/jteeuwen/go-bindata.git",
        :revision => "a0ff2567cfb70903282db057e799fd826784d41d"
  end

  def install
    ENV["GOPATH"] = buildpath
    ENV["GLIDE_HOME"] = buildpath/"glide_home"

    pkgpath = buildpath/"src/github.com/{{ .Owner }}/{{ .Name }}"
    pkgpath.install Dir["*"]
    Language::Go.stage_deps resources, buildpath/"src"

    cd "src/github.com/jteeuwen/go-bindata" do
      system "go", "install", "github.com/jteeuwen/go-bindata/..."
    end
    ENV.prepend_path "PATH", buildpath/"bin"

    cd pkgpath do
      arch = MacOS.prefer_64_bit? ? "amd64" : "386"
      ENV.deparallelize do
        system "make", "binary-darwin-#{arch}", "VERSION={{ .Tag }}", "BYPASS_GO_CHECK=yes"
      end

      bin.install "builds/bin/{{ .Tag }}/darwin/#{arch}/{{ .BinName }}"
    end
  end
end
`))

type formulaArgs struct {
	Description string
	Homepage    string
	Owner       string
	Name        string
	Tag         string
	Checksum    string
	FormulaName string
	BinName     string

	BottleURL string
	Bottles   []formulaBottleArgs
}

type formulaBottleArgs struct {
	Checksum string
	Name     string
}
