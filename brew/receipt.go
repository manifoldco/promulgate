package brew

import "text/template"

var receipt = template.Must(template.New("INSTALL_RECEIPT").Parse(`
{
   "poured_from_bottle" : false,
   "HEAD" : null,
   "source" : {
      "tap" : "{{ .Owner }}/brew",
      "path" : "@@HOMEBREW_REPOSITORY@@/Library/Taps/{{ .Owner }}/homebrew-brew/Formula/{{ .Name }}.rb",
      "versions" : {
         "devel" : null,
         "stable" : "{{ .Tag }}",
         "version_scheme" : 0,
         "head" : "HEAD"
      },
      "spec" : "stable"
   },
   "time" : null,
   "built_as_bottle" : true,
   "used_options" : [],
   "compiler" : "clang",
   "stdlib" : null,
   "unused_options" : [],
   "changed_files" : [
      "INSTALL_RECEIPT.json"
   ],
   "source_modified_time" : {{ .Time }}
}
`))

type receiptArgs struct {
	Owner string
	Name  string
	Tag   string
	Time  int64
}
