// Initially forked from http://github.com/nikhilm/gocco

package gocco

import (
	"bytes"
	"container/list"
	"github.com/russross/blackfriday"
	"path/filepath"
	"regexp"
	"html/template"
)

// ## Types
// Due to Go's statically typed nature, what is passed around in object
// literals in Docco, requires various structures

// A `Section` captures a piece of documentation and code
// Every time interleaving code is found between two comments
// a new `Section` is created.
type Section struct {
	docsText []byte
	codeText []byte
	DocsHTML []byte
	CodeHTML []byte
}

// a `TemplateSection` is a section that can be passed
// to Go's templating system, which expects strings.
type TemplateSection struct {
	DocsHTML template.HTML
	CodeHTML string
	// The `Index` field is used to create anchors to sections
	Index int
}

// a `Language` describes a programming language
type Language struct {
	// the `Pygments` name of the language
	name string
	// The comment delimiter
	symbol string
	// The regular expression to match the comment delimiter
	commentMatcher *regexp.Regexp
	// Used as a placeholder so we can parse back Pygments output
	// and put the sections together
	dividerText string
	// The HTML equivalent
	dividerHTML *regexp.Regexp
}

// a `TemplateData` is per-file
type TemplateData struct {
	// Title of the HTML output
	Title string
	// The Sections making up this file
	Sections []*TemplateSection
}

type SourceFile struct {
	Path    string
	Content []byte
}

// a map of all the languages we know
var languages map[string]*Language

// ## Main documentation generation functions

// Parse splits code into `Section`s
func parse(source string, code []byte) *list.List {
	lines := bytes.Split(code, []byte("\n"))
	sections := new(list.List)
	sections.Init()
	language := getLanguage(source)

	var hasCode bool
	var codeText = new(bytes.Buffer)
	var docsText = new(bytes.Buffer)

	// save a new section
	save := func(docs, code []byte) {
		// deep copy the slices since slices always refer to the same storage
		// by default
		docsCopy, codeCopy := make([]byte, len(docs)), make([]byte, len(code))
		copy(docsCopy, docs)
		copy(codeCopy, code)
		sections.PushBack(&Section{docsCopy, codeCopy, nil, nil})
	}

	for _, line := range lines {
		// if the line is a comment
		if language.commentMatcher.Match(line) {
			// but there was previous code
			if hasCode {
				// we need to save the existing documentation and text
				// as a section and start a new section since code blocks
				// have to be delimited before being sent to Pygments
				save(docsText.Bytes(), codeText.Bytes())
				hasCode = false
				codeText.Reset()
				docsText.Reset()
			}
			docsText.Write(language.commentMatcher.ReplaceAll(line, nil))
			docsText.WriteString("\n")
		} else {
			hasCode = true
			codeText.Write(line)
			codeText.WriteString("\n")
		}
	}
	// save any remaining parts of the source file
	save(docsText.Bytes(), codeText.Bytes())
	return sections
}

// `highlight` pipes the source to Pygments, section by section
// delimited by dividerText, then reads back the highlighted output,
// searches for the delimiters and extracts the HTML version of the code
// and documentation for each `Section`
func highlight(source string, sections *list.List) {
	//language := getLanguage(source)
	for e := sections.Front(); e != nil; e = e.Next() {
		e.Value.(*Section).CodeHTML = e.Value.(*Section).codeText
		e.Value.(*Section).DocsHTML = blackfriday.MarkdownCommon(e.Value.(*Section).docsText)
	}
}

// render the final HTML
func generateHTML(source string, sections *list.List, tpl *template.Template) []byte {
	title := filepath.Base(source)
	// convert every `Section` into corresponding `TemplateSection`
	sectionsArray := make([]*TemplateSection, sections.Len())
	for e, i := sections.Front(), 0; e != nil; e, i = e.Next(), i+1 {
		var sec = e.Value.(*Section)
		docsBuf := bytes.NewBuffer(sec.DocsHTML)
		codeBuf := bytes.NewBuffer(sec.CodeHTML)
		sectionsArray[i] = &TemplateSection{template.HTML(docsBuf.String()), codeBuf.String(), i + 1}
	}
	// run through the Go template
	html := goccoTemplate(TemplateData{title, sectionsArray}, tpl)
	return html
}

func goccoTemplate(data TemplateData, tpl *template.Template) []byte {
	buf := new(bytes.Buffer)
	err := tpl.Execute(buf, data)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// get a `Language` given a path
func getLanguage(source string) *Language {
	return languages[filepath.Ext(source)]
}

func setupLanguages() {
	languages = make(map[string]*Language)
	// you should add more languages here
	// only the first two fields should change, the rest should
	// be `nil, "", nil`
	languages[".go"] = &Language{"golang", "//", nil, "", nil}
	languages[".py"] = &Language{"python", "#", nil, "", nil}
}

func init() {
	setupLanguages()

	// create the regular expressions based on the language comment symbol
	for _, lang := range languages {
		lang.commentMatcher, _ = regexp.Compile("^\\s*" + lang.symbol + "\\s?")
		lang.dividerText = "\n" + lang.symbol + "DIVIDER\n"
		lang.dividerHTML, _ = regexp.Compile("\\n*<span class=\"c1?\">" + lang.symbol + "DIVIDER<\\/span>\\n*")
	}
}

// Generate the documentation for a single source file
// by splitting it into sections, highlighting each section
// and putting it together.
// The WaitGroup is used to signal we are done, so that the main
// goroutine waits for all the sub goroutines
func GenerateDocumentation(file *SourceFile, tpl *template.Template) []byte {
	sections := parse(file.Path, file.Content)
	highlight(file.Path, sections)
	return generateHTML(file.Path, sections, tpl)
}

// Returns true if `file` could be processed
func Allowed(file string) bool {
	ext := filepath.Ext(file)
	_, ok := languages[ext]
	return ok
}
