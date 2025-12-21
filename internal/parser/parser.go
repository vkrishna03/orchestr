package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/vkrishna03/orchestr/internal/models"
)

// Parser parses Orchestrfile content into test specifications.
type Parser struct {
	strict bool // If true, fail on unknown keywords
}

// Option configures the parser.
type Option func(*Parser)

// WithStrict enables strict mode where unknown keywords cause errors.
func WithStrict(strict bool) Option {
	return func(p *Parser) {
		p.strict = strict
	}
}

// NewParser creates a new parser with optional configuration.
func NewParser(opts ...Option) *Parser {
	p := &Parser{
		strict: false,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Parse parses Orchestrfile content from a reader.
func (p *Parser) Parse(r io.Reader) ([]models.TestSpec, error) {
	scanner := bufio.NewScanner(r)
	return p.parseLines(scanner)
}

// ParseString parses Orchestrfile content from a string.
func (p *Parser) ParseString(content string) ([]models.TestSpec, error) {
	return p.Parse(strings.NewReader(content))
}

// ParseFile parses an Orchestrfile from a file path.
func (p *Parser) ParseFile(path string) ([]models.TestSpec, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open orchestrfile: %w", err)
	}
	defer file.Close()
	return p.Parse(file)
}

// Keyword represents recognized Orchestrfile keywords.
type Keyword string

// Supported keywords.
const (
	KeywordEntity      Keyword = "ENTITY"
	KeywordEndpoint    Keyword = "ENDPOINT"
	KeywordMethod      Keyword = "METHOD"
	KeywordData        Keyword = "DATA"
	KeywordAuth        Keyword = "AUTH"
	KeywordHeaders     Keyword = "HEADERS"
	KeywordExpect      Keyword = "EXPECT"
	KeywordSave        Keyword = "SAVE"
	KeywordExpectCalls Keyword = "EXPECT_CALLS"
)

// validMethods lists allowed HTTP methods.
var validMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"PATCH":   true,
	"DELETE":  true,
	"HEAD":    true,
	"OPTIONS": true,
}

// Regex patterns for parsing.
var (
	keywordPattern    = regexp.MustCompile(`^([A-Z_]+):\s*(.*)$`)
	expectPattern     = regexp.MustCompile(`^(\d{3})(?:\s+with\s+(.*))?$`)
	savePattern       = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\s+from\s+(.+)$`)
	expectCallPattern = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*):(/?[^\s:]*)(?::([A-Z]+))?(?::(\d+))?$`)
	jsonKeyPattern    = regexp.MustCompile(`(\{|,)\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:`)
	typePatternRegex = regexp.MustCompile(`<(uuid|jwt|string|int|float|bool|email|iso8601|url|any|array|object)(:[^>]+)?>`)
)

// testCaseBuilder accumulates fields for a single test case.
type testCaseBuilder struct {
	lineNum     int
	entity      string
	endpoint    string
	method      string
	data        string
	auth        string
	expect      string
	saves       []string
	expectCalls []string
}

func newTestCaseBuilder(lineNum int) *testCaseBuilder {
	return &testCaseBuilder{lineNum: lineNum}
}

// parseLines processes lines and builds test specs.
func (p *Parser) parseLines(scanner *bufio.Scanner) ([]models.TestSpec, error) {
	var specs []models.TestSpec
	var current *testCaseBuilder
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Test case separator
		if line == "---" {
			if current != nil {
				spec, err := current.build()
				if err != nil {
					return nil, err
				}
				specs = append(specs, spec)
			}
			current = nil
			continue
		}

		// Start new test case if needed
		if current == nil {
			current = newTestCaseBuilder(lineNum)
		}

		// Parse keyword line
		if err := p.parseLine(current, line, lineNum); err != nil {
			return nil, err
		}
	}

	// Handle final test case (no trailing ---)
	if current != nil {
		spec, err := current.build()
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read orchestrfile: %w", err)
	}

	return specs, nil
}

// parseLine parses a single keyword line.
func (p *Parser) parseLine(b *testCaseBuilder, line string, lineNum int) error {
	matches := keywordPattern.FindStringSubmatch(line)
	if matches == nil {
		if p.strict {
			return &ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("invalid line format: %q", line),
			}
		}
		return nil // Ignore unrecognized lines in non-strict mode
	}

	keyword := Keyword(matches[1])
	value := strings.TrimSpace(matches[2])

	switch keyword {
	case KeywordEntity:
		b.entity = value
	case KeywordEndpoint:
		if value == "" {
			return &ParseError{Line: lineNum, Message: "ENDPOINT cannot be empty"}
		}
		b.endpoint = value
	case KeywordMethod:
		upper := strings.ToUpper(value)
		if !validMethods[upper] {
			return &ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("invalid METHOD %q, expected GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS", value),
			}
		}
		b.method = upper
	case KeywordData:
		b.data = value
	case KeywordAuth:
		b.auth = value
	case KeywordExpect:
		b.expect = value
	case KeywordSave:
		b.saves = append(b.saves, value)
	case KeywordExpectCalls:
		b.expectCalls = append(b.expectCalls, value)
	default:
		if p.strict {
			return &ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("unknown keyword %q", keyword),
			}
		}
	}

	return nil
}

// build validates and constructs a TestSpec from accumulated fields.
func (b *testCaseBuilder) build() (models.TestSpec, error) {
	// Validate required fields
	if b.endpoint == "" {
		return models.TestSpec{}, &ParseError{
			Line:    b.lineNum,
			Message: "missing required ENDPOINT",
		}
	}
	if b.method == "" {
		return models.TestSpec{}, &ParseError{
			Line:    b.lineNum,
			Message: "missing required METHOD",
		}
	}
	if b.expect == "" {
		return models.TestSpec{}, &ParseError{
			Line:    b.lineNum,
			Message: "missing required EXPECT",
		}
	}

	// Parse EXPECT
	expectSpec, err := parseExpect(b.expect)
	if err != nil {
		return models.TestSpec{}, &ParseError{
			Line:    b.lineNum,
			Message: fmt.Sprintf("invalid EXPECT: %v", err),
		}
	}

	// Parse DATA if present
	var data map[string]any
	if b.data != "" {
		data, err = parseJSON(b.data)
		if err != nil {
			return models.TestSpec{}, &ParseError{
				Line:    b.lineNum,
				Message: fmt.Sprintf("invalid DATA: %v", err),
			}
		}
	}

	// Parse SAVE directives
	var saves []models.SaveSpec
	for _, s := range b.saves {
		save, err := parseSave(s)
		if err != nil {
			return models.TestSpec{}, &ParseError{
				Line:    b.lineNum,
				Message: fmt.Sprintf("invalid SAVE: %v", err),
			}
		}
		saves = append(saves, save)
	}

	// Parse EXPECT_CALLS
	var expectCalls []models.ExpectCallSpec
	for _, ec := range b.expectCalls {
		call, err := parseExpectCall(ec)
		if err != nil {
			return models.TestSpec{}, &ParseError{
				Line:    b.lineNum,
				Message: fmt.Sprintf("invalid EXPECT_CALLS: %v", err),
			}
		}
		expectCalls = append(expectCalls, call)
	}

	return models.TestSpec{
		Entity:      b.entity,
		Endpoint:    b.endpoint,
		Method:      b.method,
		Data:        data,
		RawData:     b.data,
		Auth:        b.auth,
		Expect:      expectSpec,
		Save:        saves,
		ExpectCalls: expectCalls,
		LineNumber:  b.lineNum,
	}, nil
}

// parseExpect parses an EXPECT value like "200" or "201 with {id: <uuid>}".
func parseExpect(input string) (models.ExpectSpec, error) {
	matches := expectPattern.FindStringSubmatch(strings.TrimSpace(input))
	if matches == nil {
		return models.ExpectSpec{}, fmt.Errorf("invalid format, expected 'STATUS' or 'STATUS with {pattern}'")
	}

	status, _ := strconv.Atoi(matches[1]) // Regex ensures valid int

	spec := models.ExpectSpec{
		Status:  status,
		RawBody: matches[2],
	}

	if matches[2] != "" {
		pattern, err := parseJSON(matches[2])
		if err != nil {
			return models.ExpectSpec{}, fmt.Errorf("invalid pattern: %w", err)
		}
		spec.Pattern = pattern
	}

	return spec, nil
}

// parseJSON normalizes relaxed JSON and parses it.
// Handles both standard JSON {"key": "value"} and relaxed {key: "value"}.
func parseJSON(input string) (map[string]any, error) {
	input = strings.TrimSpace(input)

	// First, try standard JSON
	var result map[string]any
	if err := json.Unmarshal([]byte(input), &result); err == nil {
		return result, nil
	}

	// Normalize: add quotes to unquoted keys
	normalized := normalizeJSON(input)
	if err := json.Unmarshal([]byte(normalized), &result); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return result, nil
}

// normalizeJSON converts {key: value} to {"key": value} and quotes type patterns.
func normalizeJSON(input string) string {
	// Quote unquoted keys: {key: → {"key":
	result := jsonKeyPattern.ReplaceAllString(input, `$1"$2":`)
	// Quote type patterns: <uuid>, <int:gte=0>, etc. → "<uuid>", "<int:gte=0>"
	result = typePatternRegex.ReplaceAllStringFunc(result, func(match string) string {
		return `"` + match + `"`
	})
	// Quote variable references: {{var}} → "{{var}}"
	result = regexp.MustCompile(`\{\{(\w+)\}\}`).ReplaceAllString(result, `"{{$1}}"`)
	return result
}

// parseSave parses a SAVE value like "token from body.token".
func parseSave(input string) (models.SaveSpec, error) {
	matches := savePattern.FindStringSubmatch(strings.TrimSpace(input))
	if matches == nil {
		return models.SaveSpec{}, fmt.Errorf("invalid format, expected 'varname from json.path'")
	}

	return models.SaveSpec{
		VarName:  matches[1],
		JSONPath: matches[2],
	}, nil
}

// parseExpectCall parses an EXPECT_CALLS value like "stripe:/v1/charges".
func parseExpectCall(input string) (models.ExpectCallSpec, error) {
	matches := expectCallPattern.FindStringSubmatch(strings.TrimSpace(input))
	if matches == nil {
		return models.ExpectCallSpec{}, fmt.Errorf("invalid format, expected 'service:/path'")
	}

	spec := models.ExpectCallSpec{
		Service: matches[1],
		Path:    matches[2],
		Count:   1, // Default
	}

	if matches[3] != "" {
		spec.Method = matches[3]
	}
	if matches[4] != "" {
		spec.Count, _ = strconv.Atoi(matches[4])
	}

	return spec, nil
}
