package cheque

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"embed"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"project/internal/modules/banks/rocketbank/domain"
)

const (
	outputDir         = "data/reports/rocketbank/cheques"
	regularFontPath   = "internal/shared/fonts/Regular.otf"
	wideFontPath      = "internal/shared/fonts/Wide.otf"
	svgPDFRasterScale = 4
)

//go:embed templates/*.svg
var templateFS embed.FS

type Service struct{}

type chequeTemplate struct {
	Path      string
	Date      string
	Operation string
	Amount    string
	Phone     string
	Bank      string
	Client    string
	ClientTel string
	Card      string
	SBPID     string
	AuthCode  string
	RRN       string
	Document  string
}

func NewService() *Service {
	return &Service{}
}

func readChequeFont(path string) ([]byte, error) {
	candidates := []string{
		path,
		filepath.Join("../../../../../..", path),
	}
	for _, candidate := range candidates {
		body, err := os.ReadFile(candidate)
		if err == nil {
			return body, nil
		}
	}

	return nil, fmt.Errorf("read cheque font %s: file not found", path)
}

func (s *Service) GenerateMissingSBPTransferCheques(config *domain.Config) error {
	if config == nil {
		return nil
	}

	for _, item := range config.History {
		switch item.Type {
		case domain.HistoryTypeSBPTransfer:
			if err := s.GenerateSBPTransferCheque(item, config.ClientInfo); err != nil {
				return err
			}
		case domain.HistoryTypeCardTransfer:
			if err := s.GenerateCardTransferCheque(item, config.ClientInfo); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) GenerateSBPTransferCheque(item domain.HistoryItem, clientInfo domain.ClientInfo) error {
	if item.Type != domain.HistoryTypeSBPTransfer {
		return nil
	}

	tpl := outgoingTemplate()
	if domain.NormalizeHistoryDirection(item.Direction) == "INCOMING" {
		tpl = incomingTemplate()
	}

	templateBody, err := templateFS.ReadFile(tpl.Path)
	if err != nil {
		return fmt.Errorf("read cheque template: %w", err)
	}

	bank, ok := domain.SBPTransferBankByID(item.BankID)
	if !ok {
		return fmt.Errorf("generate cheque: unknown bank id %q", item.BankID)
	}

	replacements := map[string]string{
		tpl.Date:      chequeDate(item.Time),
		tpl.Operation: chequeOperationName(item),
		tpl.Amount:    chequeAmount(item.Amount),
		tpl.Phone:     formatChequePhone(item.PhoneNumber),
		tpl.Bank:      normalizeChequeText(bank.FullName),
		tpl.Client:    chequeClientName(clientInfo),
		tpl.ClientTel: formatChequePhone(stringValue(clientInfo.PhoneNumber)),
		tpl.Card:      formatChequeCardNumber(stringValue(clientInfo.CardNumber)),
		tpl.SBPID:     domain.SBPTransferOperationID(item),
		tpl.Document:  domain.HistoryItemID(item),
	}

	body, err := renderSVGTemplatePDF(templateBody, replacements)
	if err != nil {
		return err
	}

	path := SBPTransferChequePath(item)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return err
	}

	return nil
}

func (s *Service) GenerateCardTransferCheque(item domain.HistoryItem, clientInfo domain.ClientInfo) error {
	if item.Type != domain.HistoryTypeCardTransfer {
		return nil
	}

	tpl := outgoingCardTemplate()
	if domain.NormalizeHistoryDirection(item.Direction) == "INCOMING" {
		tpl = incomingCardTemplate()
	}

	templateBody, err := templateFS.ReadFile(tpl.Path)
	if err != nil {
		return fmt.Errorf("read card cheque template: %w", err)
	}

	bank, ok := domain.SBPTransferBankByID(item.BankID)
	if !ok {
		return fmt.Errorf("generate card cheque: unknown bank id %q", item.BankID)
	}

	cardNumber := item.RecipientCardNumber
	if domain.NormalizeHistoryDirection(item.Direction) == "INCOMING" {
		cardNumber = stringValue(clientInfo.CardNumber)
	}

	replacements := map[string]string{
		tpl.Date:     chequeDate(item.Time),
		tpl.Amount:   chequeAmount(item.Amount),
		tpl.Bank:     normalizeChequeText(bank.Name),
		tpl.Client:   chequeClientName(clientInfo),
		tpl.Card:     formatChequeMaskedCardNumber(cardNumber),
		tpl.SBPID:    formatChequeCardNumber(stringValue(clientInfo.CardNumber)),
		tpl.AuthCode: cardTransferAuthCode(item),
		tpl.RRN:      cardTransferRRN(item),
		tpl.Document: domain.HistoryItemID(item),
	}

	if domain.NormalizeHistoryDirection(item.Direction) == "INCOMING" {
		replacements[tpl.Bank] = "РОКЕТ"
	}

	body, err := renderSVGTemplatePDF(templateBody, replacements)
	if err != nil {
		return err
	}

	path := SBPTransferChequePath(item)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return err
	}

	return nil
}

func renderSVGTemplatePDF(templateBody []byte, replacements map[string]string) ([]byte, error) {
	svg := string(templateBody)
	for oldValue, newValue := range replacements {
		if oldValue == "" {
			continue
		}
		svg = strings.ReplaceAll(svg, oldValue, escapeSVGText(newValue))
	}
	svg = normalizeSVGChequeFontNames(svg)
	var err error
	svg, err = embedSVGChequeFonts(svg)
	if err != nil {
		return nil, err
	}
	svg = normalizeSVGChequeHeadingColor(svg)
	svg = addSVGWhiteBackground(svg)
	svg = scaleSVGCanvas(svg, svgPDFRasterScale)

	tempDir, err := os.MkdirTemp("", "rocketbank-cheque-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	svgPath := filepath.Join(tempDir, "cheque.svg")
	pdfPath := filepath.Join(tempDir, "cheque.pdf")
	if err := os.WriteFile(svgPath, []byte(svg), 0o644); err != nil {
		return nil, err
	}

	if err := convertSVGToPDF(svgPath, pdfPath); err != nil {
		return nil, err
	}

	body, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func convertSVGToPDF(svgPath string, pdfPath string) error {
	var attempts []struct {
		name string
		args []string
	}

	if runtime.GOOS == "darwin" {
		attempts = append(attempts, struct {
			name string
			args []string
		}{
			name: "sips",
			args: []string{"-s", "format", "pdf", svgPath, "--out", pdfPath},
		})
	}

	attempts = append(attempts, struct {
		name string
		args []string
	}{
		name: "rsvg-convert",
		args: []string{"-f", "pdf", "-o", pdfPath, svgPath},
	})

	var errors []string
	for _, attempt := range attempts {
		if _, err := exec.LookPath(attempt.name); err != nil {
			errors = append(errors, fmt.Sprintf("%s not found", attempt.name))
			continue
		}

		cmd := exec.Command(attempt.name, attempt.args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			errors = append(errors, fmt.Sprintf("%s failed: %v: %s", attempt.name, err, strings.TrimSpace(string(output))))
			continue
		}

		return nil
	}

	return fmt.Errorf("convert svg cheque to pdf: %s; install librsvg2-bin on Ubuntu for rsvg-convert", strings.Join(errors, "; "))
}

func normalizeSVGChequeFontNames(svg string) string {
	replacer := strings.NewReplacer(
		"font-family:RocketMono-Regular-Identity", "font-family:'Rocket Mono Regular'",
		"-inkscape-font-specification:RocketMono-Regular-Identity-H", "-inkscape-font-specification:'Rocket Mono Regular'",
		"-inkscape-font-specification:RocketMono-Regular-Identity", "-inkscape-font-specification:'Rocket Mono Regular'",
		"font-family:RocketSans-Wide-Identity", "font-family:'Rocket Sans Wide'",
		"-inkscape-font-specification:RocketSans-Wide-Identity-H", "-inkscape-font-specification:'Rocket Sans Wide'",
		"-inkscape-font-specification:RocketSans-Wide-Identity", "-inkscape-font-specification:'Rocket Sans Wide'",
	)

	return replacer.Replace(svg)
}

func embedSVGChequeFonts(svg string) (string, error) {
	if !strings.Contains(svg, "Rocket Mono Regular") && !strings.Contains(svg, "Rocket Sans Wide") {
		return svg, nil
	}
	if strings.Contains(svg, "@font-face") || strings.Contains(svg, `data-cheque-fonts="true"`) {
		return svg, nil
	}

	style, err := sbpChequeFontFaceStyle()
	if err != nil {
		return "", err
	}

	pattern := regexp.MustCompile(`(<defs[^>]*>)`)
	if pattern.MatchString(svg) {
		return pattern.ReplaceAllString(svg, "$1"+style), nil
	}

	return strings.Replace(svg, "><metadata", ">"+style+"<metadata", 1), nil
}

func sbpChequeFontFaceStyle() (string, error) {
	templateBody, err := templateFS.ReadFile(outgoingTemplate().Path)
	if err != nil {
		return "", fmt.Errorf("read sbp cheque fonts: %w", err)
	}

	pattern := regexp.MustCompile(`(?s)<style[^>]*><!\[CDATA\[(.*?)\]\]></style>`)
	match := pattern.FindStringSubmatch(string(templateBody))
	if len(match) != 2 {
		return "", fmt.Errorf("sbp cheque fonts not found")
	}

	return `<style data-cheque-fonts="true" type="text/css"><![CDATA[` + match[1] + `]]></style>`, nil
}

func normalizeSVGChequeHeadingColor(svg string) string {
	return strings.ReplaceAll(svg, "fill:#000000;fill-opacity:0.3", "fill:#b2b2b2;fill-opacity:1")
}

func addSVGWhiteBackground(svg string) string {
	if strings.Contains(svg, `data-cheque-background="true"`) {
		return svg
	}

	background := `<rect data-cheque-background="true" width="100%" height="100%" fill="#ffffff"/>`
	if strings.Contains(svg, "><metadata") {
		return strings.Replace(svg, "><metadata", ">"+background+"<metadata", 1)
	}

	return strings.Replace(svg, "><defs", ">"+background+"<defs", 1)
}

func scaleSVGCanvas(svg string, scale int) string {
	if scale <= 1 {
		return svg
	}

	svg = scaleSVGDimension(svg, "width", scale)
	svg = scaleSVGDimension(svg, "height", scale)
	return svg
}

func scaleSVGDimension(svg string, attribute string, scale int) string {
	pattern := regexp.MustCompile(attribute + `="([0-9]+)"`)
	return pattern.ReplaceAllStringFunc(svg, func(match string) string {
		parts := pattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		value, err := strconv.Atoi(parts[1])
		if err != nil {
			return match
		}
		return fmt.Sprintf(`%s="%d"`, attribute, value*scale)
	})
}

func escapeSVGText(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(value)
}

type chequeLine struct {
	Font  string
	Size  int
	X     float64
	Y     float64
	Color string
	Text  string
}

func renderChequePDF(tpl chequeTemplate, replacements map[string]string) ([]byte, error) {
	regularFont, err := readChequeFont(regularFontPath)
	if err != nil {
		return nil, fmt.Errorf("read regular cheque font: %w", err)
	}
	wideFont, err := readChequeFont(wideFontPath)
	if err != nil {
		return nil, fmt.Errorf("read wide cheque font: %w", err)
	}
	regularCMap, err := parseFontCMap(regularFont)
	if err != nil {
		return nil, fmt.Errorf("parse regular cheque font cmap: %w", err)
	}
	wideCMap, err := parseFontCMap(wideFont)
	if err != nil {
		return nil, fmt.Errorf("parse wide cheque font cmap: %w", err)
	}

	lines := chequeLines(tpl, replacements)
	pageHeight := 1116.0
	if tpl.Path == incomingTemplate().Path {
		pageHeight = 1050.0
	}

	content, regularToUnicode, wideToUnicode, err := chequeContentStream(lines, regularCMap, wideCMap)
	if err != nil {
		return nil, err
	}
	return buildChequePDF(420, pageHeight, content, regularFont, wideFont, regularToUnicode, wideToUnicode), nil
}

func chequeLines(tpl chequeTemplate, replacements map[string]string) []chequeLine {
	value := func(source string) string {
		if replacement, ok := replacements[source]; ok {
			return replacement
		}
		return source
	}
	label := func(y float64, text string) chequeLine {
		return chequeLine{Font: "Regular", Size: 14, X: 36, Y: y, Color: "0 0 0", Text: text}
	}
	field := func(y float64, text string) chequeLine {
		return chequeLine{Font: "Regular", Size: 14, X: 36, Y: y, Color: "0.06275 0.06275 0.06275", Text: value(text)}
	}
	amount := func(y float64) chequeLine {
		return chequeLine{Font: "Wide", Size: 19, X: 36, Y: y, Color: "0 0 0", Text: value(tpl.Amount)}
	}

	if tpl.Path == incomingTemplate().Path {
		return []chequeLine{
			field(935.87, tpl.Date),
			amount(914.74),
			label(875.87, "ПЕРЕВОД ЧЕРЕЗ СБП"),
			field(856.5, "ПОПОЛНЕНИЕ"),
			label(817.12, "ОТПРАВИТЕЛЬ"),
			field(797.75, tpl.Operation),
			label(758.37, "ТЕЛЕФОН ОТПРАВИТЕЛЯ"),
			field(738.99, tpl.Phone),
			label(699.62, "БАНК ОТПРАВИТЕЛЯ"),
			field(680.24, tpl.Bank),
			label(640.87, "ПОЛУЧАТЕЛЬ"),
			field(621.49, tpl.Client),
			label(582.11, "ТЕЛЕФОН ПОЛУЧАТЕЛЯ"),
			field(562.74, tpl.ClientTel),
			label(523.36, "БАНК ПОЛУЧАТЕЛЯ"),
			field(503.99, "РОКЕТ"),
			label(464.61, "СЧЁТ ПОЛУЧАТЕЛЯ"),
			field(445.23, tpl.Card),
			label(405.86, "ИДЕНТИФИКАТОР ОПЕРАЦИИ СБП"),
			field(386.48, tpl.SBPID),
			label(347.11, "СТАТУС"),
			field(327.73, "УСПЕШНО"),
			label(288.35, "НОМЕР ДОКУМЕНТА"),
			field(268.98, tpl.Document),
		}
	}

	return []chequeLine{
		field(1001.87, tpl.Date),
		amount(980.74),
		label(941.87, "КОМИССИЯ"),
		field(922.5, "БЕЗ КОМИССИИ"),
		label(883.12, "ПЕРЕВОД ЧЕРЕЗ СБП"),
		field(863.75, "ИСХОДЯЩИЙ"),
		label(824.37, "ПОЛУЧАТЕЛЬ"),
		field(804.99, tpl.Operation),
		label(765.62, "ТЕЛЕФОН ПОЛУЧАТЕЛЯ"),
		field(746.24, tpl.Phone),
		label(706.87, "БАНК ПОЛУЧАТЕЛЯ"),
		field(687.49, tpl.Bank),
		label(648.11, "ОТПРАВИТЕЛЬ"),
		field(628.74, tpl.Client),
		label(589.36, "ТЕЛЕФОН ОТПРАВИТЕЛЯ"),
		field(569.99, tpl.ClientTel),
		label(530.61, "БАНК ОТПРАВИТЕЛЯ"),
		field(511.23, "РОКЕТ"),
		label(471.86, "СЧЁТ ОТПРАВИТЕЛЯ"),
		field(452.48, tpl.Card),
		label(413.11, "ИДЕНТИФИКАТОР ОПЕРАЦИИ СБП"),
		field(393.73, tpl.SBPID),
		label(354.35, "СТАТУС"),
		field(334.98, "УСПЕШНО"),
		label(295.6, "НОМЕР ДОКУМЕНТА"),
		field(276.23, tpl.Document),
	}
}

func chequeContentStream(lines []chequeLine, regularCMap map[rune]uint16, wideCMap map[rune]uint16) ([]byte, map[uint16]rune, map[uint16]rune, error) {
	var content bytes.Buffer
	regularToUnicode := map[uint16]rune{}
	wideToUnicode := map[uint16]rune{}
	for _, line := range lines {
		fontName := "FRegular"
		cmap := regularCMap
		toUnicode := regularToUnicode
		if line.Font == "Wide" {
			fontName = "FWide"
			cmap = wideCMap
			toUnicode = wideToUnicode
		}
		hexText, err := encodeFontGlyphs(line.Text, cmap, toUnicode)
		if err != nil {
			return nil, nil, nil, err
		}
		content.WriteString("BT\n")
		content.WriteString(fmt.Sprintf("/%s %d Tf\n", fontName, line.Size))
		content.WriteString(fmt.Sprintf("%.2f %.2f Td\n", line.X, line.Y))
		content.WriteString(line.Color + " rg\n")
		content.WriteString("<" + hexText + "> Tj\n")
		content.WriteString("ET\n")
	}

	return content.Bytes(), regularToUnicode, wideToUnicode, nil
}

func buildChequePDF(width float64, height float64, content []byte, regularFont []byte, wideFont []byte, regularToUnicode map[uint16]rune, wideToUnicode map[uint16]rune) []byte {
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.0f %.0f] /Resources << /Font << /FRegular 6 0 R /FWide 9 0 R >> >> /Contents 4 0 R >>", width, height),
		streamObject(content),
		fontStreamObject(regularFont),
		"<< /Type /Font /Subtype /Type0 /BaseFont /RocketSans-Regular /Encoding /Identity-H /DescendantFonts [7 0 R] /ToUnicode 11 0 R >>",
		"<< /Type /Font /Subtype /CIDFontType0 /BaseFont /RocketSans-Regular /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> /CIDToGIDMap /Identity /FontDescriptor 8 0 R /DW 600 >>",
		"<< /Type /FontDescriptor /FontName /RocketSans-Regular /Flags 4 /FontBBox [0 -250 1000 1000] /ItalicAngle 0 /Ascent 900 /Descent -250 /CapHeight 700 /StemV 80 /FontFile3 5 0 R >>",
		"<< /Type /Font /Subtype /Type0 /BaseFont /RocketSans-Wide /Encoding /Identity-H /DescendantFonts [10 0 R] /ToUnicode 11 0 R >>",
		"<< /Type /Font /Subtype /CIDFontType0 /BaseFont /RocketSans-Wide /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> /CIDToGIDMap /Identity /FontDescriptor 12 0 R /DW 700 >>",
		streamObject([]byte(toUnicodeCMap(regularToUnicode))),
		"<< /Type /FontDescriptor /FontName /RocketSans-Wide /Flags 4 /FontBBox [0 -250 1000 1000] /ItalicAngle 0 /Ascent 900 /Descent -250 /CapHeight 700 /StemV 80 /FontFile3 13 0 R >>",
		fontStreamObject(wideFont),
		streamObject([]byte(toUnicodeCMap(wideToUnicode))),
	}
	objects[8] = "<< /Type /Font /Subtype /Type0 /BaseFont /RocketSans-Wide /Encoding /Identity-H /DescendantFonts [10 0 R] /ToUnicode 14 0 R >>"

	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.7\n%\xE2\xE3\xCF\xD3\n")
	offsets := make([]int, len(objects)+1)
	for index, object := range objects {
		objectNumber := index + 1
		offsets[objectNumber] = pdf.Len()
		pdf.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", objectNumber, object))
	}
	xrefOffset := pdf.Len()
	pdf.WriteString(fmt.Sprintf("xref\n0 %d\n", len(objects)+1))
	pdf.WriteString("0000000000 65535 f \n")
	for objectNumber := 1; objectNumber <= len(objects); objectNumber++ {
		pdf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[objectNumber]))
	}
	pdf.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xrefOffset))

	return pdf.Bytes()
}

func streamObject(body []byte) string {
	return fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(body), string(body))
}

func fontStreamObject(body []byte) string {
	return fmt.Sprintf("<< /Length %d /Subtype /OpenType >>\nstream\n%s\nendstream", len(body), string(body))
}

func encodeFontGlyphs(value string, cmap map[rune]uint16, toUnicode map[uint16]rune) (string, error) {
	raw := make([]byte, 0, len([]rune(value))*2)
	for _, char := range value {
		glyphID, ok := cmap[char]
		if !ok {
			return "", fmt.Errorf("font glyph not found for %q in %q", char, value)
		}
		toUnicode[glyphID] = char
		raw = append(raw, byte(glyphID>>8), byte(glyphID))
	}

	return strings.ToUpper(hex.EncodeToString(raw)), nil
}

func toUnicodeCMap(mapping map[uint16]rune) string {
	var builder strings.Builder
	builder.WriteString(`/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def
/CMapName /Adobe-Identity-UCS def
/CMapType 2 def
1 begincodespacerange
<0000> <FFFF>
endcodespacerange
`)
	builder.WriteString(fmt.Sprintf("%d beginbfchar\n", len(mapping)))
	for glyphID, char := range mapping {
		builder.WriteString(fmt.Sprintf("<%04X> <%04X>\n", glyphID, char))
	}
	builder.WriteString(`endbfchar
endcmap
CMapName currentdict /CMap defineresource pop
end
end`)

	return builder.String()
}

func parseFontCMap(font []byte) (map[rune]uint16, error) {
	tables, err := fontTables(font)
	if err != nil {
		return nil, err
	}
	cmapTable, ok := tables["cmap"]
	if !ok {
		return nil, fmt.Errorf("cmap table not found")
	}
	if len(cmapTable) < 4 {
		return nil, fmt.Errorf("cmap table too short")
	}
	numTables := int(binary.BigEndian.Uint16(cmapTable[2:4]))
	type cmapRecord struct {
		platformID uint16
		encodingID uint16
		offset     uint32
		format     uint16
	}
	records := make([]cmapRecord, 0, numTables)
	for index := 0; index < numTables; index++ {
		recordOffset := 4 + index*8
		if recordOffset+8 > len(cmapTable) {
			return nil, fmt.Errorf("cmap record out of range")
		}
		offset := binary.BigEndian.Uint32(cmapTable[recordOffset+4 : recordOffset+8])
		if int(offset)+2 > len(cmapTable) {
			continue
		}
		records = append(records, cmapRecord{
			platformID: binary.BigEndian.Uint16(cmapTable[recordOffset : recordOffset+2]),
			encodingID: binary.BigEndian.Uint16(cmapTable[recordOffset+2 : recordOffset+4]),
			offset:     offset,
			format:     binary.BigEndian.Uint16(cmapTable[offset : offset+2]),
		})
	}
	for _, record := range records {
		if record.format == 12 {
			return parseCMapFormat12(cmapTable[record.offset:])
		}
	}
	for _, record := range records {
		if record.format == 4 {
			return parseCMapFormat4(cmapTable[record.offset:])
		}
	}

	return nil, fmt.Errorf("supported cmap format not found")
}

func fontTables(font []byte) (map[string][]byte, error) {
	if len(font) < 12 {
		return nil, fmt.Errorf("font too short")
	}
	numTables := int(binary.BigEndian.Uint16(font[4:6]))
	tables := map[string][]byte{}
	for index := 0; index < numTables; index++ {
		recordOffset := 12 + index*16
		if recordOffset+16 > len(font) {
			return nil, fmt.Errorf("font table record out of range")
		}
		tag := string(font[recordOffset : recordOffset+4])
		offset := int(binary.BigEndian.Uint32(font[recordOffset+8 : recordOffset+12]))
		length := int(binary.BigEndian.Uint32(font[recordOffset+12 : recordOffset+16]))
		if offset < 0 || length < 0 || offset+length > len(font) {
			return nil, fmt.Errorf("font table %s out of range", tag)
		}
		tables[tag] = font[offset : offset+length]
	}

	return tables, nil
}

func parseCMapFormat12(table []byte) (map[rune]uint16, error) {
	if len(table) < 16 {
		return nil, fmt.Errorf("cmap format 12 too short")
	}
	groups := int(binary.BigEndian.Uint32(table[12:16]))
	result := map[rune]uint16{}
	for index := 0; index < groups; index++ {
		offset := 16 + index*12
		if offset+12 > len(table) {
			return nil, fmt.Errorf("cmap format 12 group out of range")
		}
		startChar := binary.BigEndian.Uint32(table[offset : offset+4])
		endChar := binary.BigEndian.Uint32(table[offset+4 : offset+8])
		startGlyph := binary.BigEndian.Uint32(table[offset+8 : offset+12])
		for char := startChar; char <= endChar; char++ {
			result[rune(char)] = uint16(startGlyph + char - startChar)
		}
	}

	return result, nil
}

func parseCMapFormat4(table []byte) (map[rune]uint16, error) {
	if len(table) < 16 {
		return nil, fmt.Errorf("cmap format 4 too short")
	}
	segCount := int(binary.BigEndian.Uint16(table[6:8]) / 2)
	endCodesOffset := 14
	startCodesOffset := endCodesOffset + segCount*2 + 2
	idDeltaOffset := startCodesOffset + segCount*2
	idRangeOffsetOffset := idDeltaOffset + segCount*2
	if idRangeOffsetOffset+segCount*2 > len(table) {
		return nil, fmt.Errorf("cmap format 4 arrays out of range")
	}

	result := map[rune]uint16{}
	for segment := 0; segment < segCount; segment++ {
		endCode := binary.BigEndian.Uint16(table[endCodesOffset+segment*2 : endCodesOffset+segment*2+2])
		startCode := binary.BigEndian.Uint16(table[startCodesOffset+segment*2 : startCodesOffset+segment*2+2])
		idDelta := int16(binary.BigEndian.Uint16(table[idDeltaOffset+segment*2 : idDeltaOffset+segment*2+2]))
		idRangeOffset := binary.BigEndian.Uint16(table[idRangeOffsetOffset+segment*2 : idRangeOffsetOffset+segment*2+2])
		if startCode == 0xFFFF && endCode == 0xFFFF {
			continue
		}
		for char := startCode; char <= endCode; char++ {
			var glyphID uint16
			if idRangeOffset == 0 {
				glyphID = uint16(int(char)+int(idDelta)) & 0xFFFF
			} else {
				glyphIndexOffset := idRangeOffsetOffset + segment*2 + int(idRangeOffset) + int(char-startCode)*2
				if glyphIndexOffset+2 > len(table) {
					continue
				}
				glyphID = binary.BigEndian.Uint16(table[glyphIndexOffset : glyphIndexOffset+2])
				if glyphID != 0 {
					glyphID = uint16(int(glyphID)+int(idDelta)) & 0xFFFF
				}
			}
			if glyphID != 0 {
				result[rune(char)] = glyphID
			}
		}
	}

	return result, nil
}

func SBPTransferChequePath(item domain.HistoryItem) string {
	return filepath.Join(outputDir, domain.HistoryItemID(item)+".pdf")
}

func outgoingTemplate() chequeTemplate {
	return chequeTemplate{
		Path:      "templates/outgoing_sbp_template.svg",
		Date:      "16.05.2026 18:14 ПО МСК",
		Operation: "АЗАТ АЛИКОВИЧ Г",
		Amount:    "50 ₽",
		Phone:     "+7 909 933-40-05",
		Bank:      `АО "ТБАНК"`,
		Client:    "МАКСИМ АЛЕКСАНДРОВИЧ Н.",
		ClientTel: "+7 983 543-99-99",
		Card:      "40817 81035 02245 32469",
		SBPID:     "B61361514043330I0B10100011760501",
		Document:  "M70093717871",
	}
}

func incomingTemplate() chequeTemplate {
	return chequeTemplate{
		Path:      "templates/incoming_sbp_template.svg",
		Date:      "16.05.2026 18:10 ПО МСК",
		Operation: "АЗАТ АЛИКОВИЧ Г",
		Amount:    "50 ₽",
		Phone:     "+7 909 933-40-05",
		Bank:      `АО "ТБАНК"`,
		Client:    "МАКСИМ АЛЕКСАНДРОВИЧ Н.",
		ClientTel: "+7 983 543-99-99",
		Card:      "40817 81035 02245 32469",
		SBPID:     "A61361510392060D0B10080011760501",
		Document:  "M70093674996",
	}
}

func outgoingCardTemplate() chequeTemplate {
	return chequeTemplate{
		Path:     "templates/outgoing_card_template.svg",
		Date:     "16.05.2026 18:15 ПО МСК",
		Amount:   "50 ₽",
		Bank:     "Т-БАНК",
		Client:   "МАКСИМ АЛЕКСАНДРОВИЧ Н.",
		Card:     "2200 **** **** 6863",
		SBPID:    "40817 81035 02245 32469",
		Document: "M70093730751",
	}
}

func incomingCardTemplate() chequeTemplate {
	return chequeTemplate{
		Path:     "templates/incoming_card_template.svg",
		Date:     "16.05.2026 18:11 ПО МСК",
		Amount:   "50 ₽",
		Bank:     "РОКЕТ",
		Card:     "2200 **** **** 1928",
		AuthCode: "386393",
		RRN:      "613615859247",
		Document: "T11650918620",
	}
}

func replaceTemplateText(templateBody []byte, replacements map[string]string) ([]byte, error) {
	streamInfo, err := firstFlateStream(templateBody)
	if err != nil {
		return nil, err
	}

	decoded, err := zlib.NewReader(bytes.NewReader(streamInfo.Body))
	if err != nil {
		return nil, err
	}
	contentBody, err := io.ReadAll(decoded)
	_ = decoded.Close()
	if err != nil {
		return nil, err
	}

	customFonts := customFontUsage{
		RegularToUnicode: map[uint16]rune{},
		WideToUnicode:    map[uint16]rune{},
	}
	content := string(contentBody)
	for oldValue, newValue := range replacements {
		oldHex, err := encodeChequeText(oldValue)
		if err != nil {
			return nil, err
		}
		newHex, err := encodeChequeText(newValue)
		if err != nil {
			var customHex string
			customHex, customFonts, err = encodeCustomChequeText(newValue, customFonts)
			if err != nil {
				return nil, err
			}
			content, err = replaceTextObjectWithCustomFont(content, oldHex, customHex, customFonts.FontResource)
			if err != nil {
				return nil, err
			}
			continue
		}
		if !strings.Contains(content, oldHex) {
			return nil, fmt.Errorf("cheque template field not found: %q", oldValue)
		}
		content = strings.Replace(content, oldHex, newHex, 1)
	}

	var compressed bytes.Buffer
	writer, err := zlib.NewWriterLevel(&compressed, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := writer.Write([]byte(content)); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	output := append([]byte(nil), templateBody...)
	if compressed.Len() > len(streamInfo.Body) || customFonts.Needed() {
		output, err = rebuildPDFWithStream(output, streamInfo.ObjectNumber, streamInfo.Generation, compressed.Bytes())
		if err != nil {
			return nil, err
		}
		if customFonts.Needed() {
			return injectCustomChequeFonts(output, customFonts)
		}
		return output, nil
	}

	copy(output[streamInfo.Start:streamInfo.End], compressed.Bytes())
	for index := streamInfo.Start + compressed.Len(); index < streamInfo.End; index++ {
		output[index] = ' '
	}

	return output, nil
}

type customFontUsage struct {
	UseRegular       bool
	UseWide          bool
	FontResource     string
	RegularToUnicode map[uint16]rune
	WideToUnicode    map[uint16]rune
	regularCMap      map[rune]uint16
	wideCMap         map[rune]uint16
}

func (usage customFontUsage) Needed() bool {
	return usage.UseRegular || usage.UseWide
}

func encodeCustomChequeText(value string, usage customFontUsage) (string, customFontUsage, error) {
	useWide := strings.Contains(value, "₽")
	if useWide {
		if usage.wideCMap == nil {
			font, err := readChequeFont(wideFontPath)
			if err != nil {
				return "", usage, err
			}
			cmap, err := parseFontCMap(font)
			if err != nil {
				return "", usage, err
			}
			usage.wideCMap = cmap
		}
		usage.UseWide = true
		usage.FontResource = "FCW"
		encoded, err := encodeFontGlyphs(value, usage.wideCMap, usage.WideToUnicode)
		return encoded, usage, err
	}

	if usage.regularCMap == nil {
		font, err := readChequeFont(regularFontPath)
		if err != nil {
			return "", usage, err
		}
		cmap, err := parseFontCMap(font)
		if err != nil {
			return "", usage, err
		}
		usage.regularCMap = cmap
	}
	usage.UseRegular = true
	usage.FontResource = "FCR"
	encoded, err := encodeFontGlyphs(value, usage.regularCMap, usage.RegularToUnicode)
	return encoded, usage, err
}

func replaceTextObjectWithCustomFont(content string, oldHex string, newHex string, fontResource string) (string, error) {
	pattern := regexp.MustCompile(`(?s)/F\d+\s+([0-9.]+)\s+Tf(.*?` + regexp.QuoteMeta(oldHex) + `)`)
	match := pattern.FindStringSubmatchIndex(content)
	if match == nil {
		return "", fmt.Errorf("cheque template field not found for custom font")
	}

	replacement := "/" + fontResource + " " + content[match[2]:match[3]] + " Tf" + strings.Replace(content[match[4]:match[5]], oldHex, "<"+newHex+">Tj", 1)
	return content[:match[0]] + replacement + content[match[1]:], nil
}

func injectCustomChequeFonts(body []byte, usage customFontUsage) ([]byte, error) {
	startXref, err := lastStartXref(body)
	if err != nil {
		return nil, err
	}
	root, info, id, _, err := lastTrailerValues(body)
	if err != nil {
		return nil, err
	}

	prefix := append([]byte(nil), body[:startXref]...)
	nextObject := maxPDFObjectNumber(prefix) + 1
	fontResources := map[string]int{}
	var objects bytes.Buffer

	if usage.UseRegular {
		font, err := readChequeFont(regularFontPath)
		if err != nil {
			return nil, err
		}
		type0Object := nextObject
		fontResources["FCR"] = type0Object
		nextObject = appendCustomFontObjects(&objects, nextObject, "RocketSans-Regular-Custom", font, usage.RegularToUnicode, 600)
	}
	if usage.UseWide {
		font, err := readChequeFont(wideFontPath)
		if err != nil {
			return nil, err
		}
		type0Object := nextObject
		fontResources["FCW"] = type0Object
		nextObject = appendCustomFontObjects(&objects, nextObject, "RocketSans-Wide-Custom", font, usage.WideToUnicode, 700)
	}

	updatedPrefix, err := injectPageFontResources(prefix, fontResources)
	if err != nil {
		return nil, err
	}
	if len(updatedPrefix) > 0 && updatedPrefix[len(updatedPrefix)-1] != '\n' {
		updatedPrefix = append(updatedPrefix, '\n')
	}
	updatedPrefix = append(updatedPrefix, objects.Bytes()...)

	return appendFreshXref(updatedPrefix, root, info, id), nil
}

func appendCustomFontObjects(objects *bytes.Buffer, startObject int, fontName string, font []byte, toUnicode map[uint16]rune, defaultWidth int) int {
	type0Object := startObject
	cidObject := startObject + 1
	descriptorObject := startObject + 2
	fontFileObject := startObject + 3
	toUnicodeObject := startObject + 4

	writePDFObject(objects, type0Object, fmt.Sprintf("<< /Type /Font /Subtype /Type0 /BaseFont /%s /Encoding /Identity-H /DescendantFonts [%d 0 R] /ToUnicode %d 0 R >>", fontName, cidObject, toUnicodeObject))
	writePDFObject(objects, cidObject, fmt.Sprintf("<< /Type /Font /Subtype /CIDFontType0 /BaseFont /%s /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> /CIDToGIDMap /Identity /FontDescriptor %d 0 R /DW %d >>", fontName, descriptorObject, defaultWidth))
	writePDFObject(objects, descriptorObject, fmt.Sprintf("<< /Type /FontDescriptor /FontName /%s /Flags 4 /FontBBox [0 -250 1000 1000] /ItalicAngle 0 /Ascent 900 /Descent -250 /CapHeight 700 /StemV 80 /FontFile3 %d 0 R >>", fontName, fontFileObject))
	writePDFObject(objects, fontFileObject, fontStreamObject(font))
	writePDFObject(objects, toUnicodeObject, streamObject([]byte(toUnicodeCMap(toUnicode))))

	return startObject + 5
}

func writePDFObject(objects *bytes.Buffer, objectNumber int, body string) {
	objects.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", objectNumber, body))
}

func injectPageFontResources(body []byte, fontResources map[string]int) ([]byte, error) {
	if len(fontResources) == 0 {
		return body, nil
	}

	var additions strings.Builder
	for name, objectNumber := range fontResources {
		additions.WriteString(fmt.Sprintf("/%s %d 0 R", name, objectNumber))
	}

	pattern := regexp.MustCompile(`/Font\s*<<`)
	location := pattern.FindIndex(body)
	if location == nil {
		return nil, fmt.Errorf("page font resources not found")
	}

	output := append([]byte(nil), body[:location[1]]...)
	output = append(output, []byte(additions.String())...)
	output = append(output, body[location[1]:]...)
	return output, nil
}

func maxPDFObjectNumber(body []byte) int {
	maxObject := 0
	objectPattern := regexp.MustCompile(`(?m)^(\d+)\s+\d+\s+obj\b`)
	matches := objectPattern.FindAllSubmatch(body, -1)
	for _, match := range matches {
		objectNumber, err := strconv.Atoi(string(match[1]))
		if err == nil && objectNumber > maxObject {
			maxObject = objectNumber
		}
	}

	return maxObject
}

type flateStreamInfo struct {
	ObjectNumber int
	Generation   int
	Start        int
	End          int
	Body         []byte
}

func firstFlateStream(body []byte) (flateStreamInfo, error) {
	streamPattern := regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj\s*<<.*?/Filter/FlateDecode/Length\s+\d+>>stream\r?\n`)
	match := streamPattern.FindSubmatchIndex(body)
	if match == nil {
		return flateStreamInfo{}, fmt.Errorf("flate stream not found")
	}

	objectNumber, err := strconv.Atoi(string(body[match[2]:match[3]]))
	if err != nil {
		return flateStreamInfo{}, err
	}
	generation, err := strconv.Atoi(string(body[match[4]:match[5]]))
	if err != nil {
		return flateStreamInfo{}, err
	}

	streamStart := match[1]
	endMarker := []byte("\nendstream")
	relativeEnd := bytes.Index(body[streamStart:], endMarker)
	if relativeEnd == -1 {
		return flateStreamInfo{}, fmt.Errorf("stream end not found")
	}

	streamEnd := streamStart + relativeEnd
	stream := bytes.TrimRight(body[streamStart:streamEnd], "\r\n")

	return flateStreamInfo{
		ObjectNumber: objectNumber,
		Generation:   generation,
		Start:        streamStart,
		End:          streamStart + len(stream),
		Body:         stream,
	}, nil
}

func rebuildPDFWithStream(body []byte, objectNumber int, generation int, stream []byte) ([]byte, error) {
	startXref, err := lastStartXref(body)
	if err != nil {
		return nil, err
	}
	root, info, id, _, err := lastTrailerValues(body)
	if err != nil {
		return nil, err
	}

	prefix := append([]byte(nil), body[:startXref]...)
	objectPattern := regexp.MustCompile(fmt.Sprintf(`(?ms)^%d\s+%d\s+obj\b.*?^endobj\s*`, objectNumber, generation))
	match := objectPattern.FindIndex(prefix)
	if match == nil {
		return nil, fmt.Errorf("pdf object %d %d not found", objectNumber, generation)
	}

	var replacement bytes.Buffer
	replacement.WriteString(fmt.Sprintf("%d %d obj\n<</Filter/FlateDecode/Length %d>>stream\n", objectNumber, generation, len(stream)))
	replacement.Write(stream)
	replacement.WriteString("\nendstream\nendobj\n")

	rebuilt := make([]byte, 0, len(prefix)-match[1]+match[0]+replacement.Len()+4096)
	rebuilt = append(rebuilt, prefix[:match[0]]...)
	rebuilt = append(rebuilt, replacement.Bytes()...)
	rebuilt = append(rebuilt, prefix[match[1]:]...)

	return appendFreshXref(rebuilt, root, info, id), nil
}

func appendFreshXref(body []byte, root string, info string, id string) []byte {
	objectOffsets := map[int]int{}
	objectGenerations := map[int]int{}
	objectPattern := regexp.MustCompile(`(?m)^(\d+)\s+(\d+)\s+obj\b`)
	matches := objectPattern.FindAllSubmatchIndex(body, -1)
	maxObject := 0
	for _, match := range matches {
		objectNumber, _ := strconv.Atoi(string(body[match[2]:match[3]]))
		generation, _ := strconv.Atoi(string(body[match[4]:match[5]]))
		objectOffsets[objectNumber] = match[0]
		objectGenerations[objectNumber] = generation
		if objectNumber > maxObject {
			maxObject = objectNumber
		}
	}

	size := maxObject + 1
	xrefOffset := len(body)
	var xref bytes.Buffer
	if len(body) > 0 && body[len(body)-1] != '\n' {
		xref.WriteByte('\n')
		xrefOffset++
	}
	xref.WriteString(fmt.Sprintf("xref\n0 %d\n", size))
	xref.WriteString("0000000000 65535 f \n")
	for objectNumber := 1; objectNumber < size; objectNumber++ {
		offset, ok := objectOffsets[objectNumber]
		if !ok {
			xref.WriteString("0000000000 65535 f \n")
			continue
		}
		xref.WriteString(fmt.Sprintf("%010d %05d n \n", offset, objectGenerations[objectNumber]))
	}

	xref.WriteString("trailer\n<<")
	xref.WriteString(fmt.Sprintf("/Size %d/Root %s", size, root))
	if info != "" {
		xref.WriteString("/Info " + info)
	}
	if id != "" {
		xref.WriteString("/ID " + id)
	}
	xref.WriteString(">>\n")
	xref.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF\n", xrefOffset))

	return append(body, xref.Bytes()...)
}

func lastStartXref(body []byte) (int, error) {
	matches := regexp.MustCompile(`startxref\s+(\d+)\s+%%EOF`).FindAllSubmatch(body, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("startxref not found")
	}

	return strconv.Atoi(string(matches[len(matches)-1][1]))
}

func lastTrailerValues(body []byte) (string, string, string, int, error) {
	matches := regexp.MustCompile(`(?s)trailer\s*<<(.*?)>>`).FindAllSubmatch(body, -1)
	if len(matches) == 0 {
		return "", "", "", 0, fmt.Errorf("trailer not found")
	}
	trailer := matches[len(matches)-1][1]

	rootMatch := regexp.MustCompile(`/Root\s+(\d+\s+\d+\s+R)`).FindSubmatch(trailer)
	if rootMatch == nil {
		return "", "", "", 0, fmt.Errorf("trailer root not found")
	}
	sizeMatch := regexp.MustCompile(`/Size\s+(\d+)`).FindSubmatch(trailer)
	if sizeMatch == nil {
		return "", "", "", 0, fmt.Errorf("trailer size not found")
	}
	size, err := strconv.Atoi(string(sizeMatch[1]))
	if err != nil {
		return "", "", "", 0, err
	}
	info := ""
	if infoMatch := regexp.MustCompile(`/Info\s+(\d+\s+\d+\s+R)`).FindSubmatch(trailer); infoMatch != nil {
		info = string(infoMatch[1])
	}
	id := ""
	if idMatch := regexp.MustCompile(`(?s)/ID\s+(\[<.*?>\s*<.*?>\])`).FindSubmatch(trailer); idMatch != nil {
		id = string(idMatch[1])
	}

	return string(rootMatch[1]), info, id, size, nil
}

func encodeChequeText(value string) (string, error) {
	value = normalizeChequeText(value)
	encoder := chequeRegularGlyphs
	if strings.Contains(value, "₽") {
		encoder = chequeWideGlyphs
	}

	var builder strings.Builder
	builder.WriteString("<")
	for _, char := range value {
		code, ok := encoder[char]
		if !ok {
			return "", fmt.Errorf("cheque font glyph not found for %q in %q", char, value)
		}
		builder.WriteString(code)
	}
	builder.WriteString(">Tj")

	return builder.String(), nil
}

var chequeRegularGlyphs = map[rune]string{
	' ': "0062", '"': "002f", '+': "0043", '-': "000e", '.': "0035", ':': "0037",
	'0': "0003", '1': "0004", '2': "0005", '3': "0006", '4': "0007",
	'5': "0008", '6': "0009", '7': "000a", '8': "000b", '9': "000c",
	'A': "0068", 'B': "0069", 'D': "006b", 'I': "0070", 'M': "0074",
	'Ё': "00a9", 'А': "00aa", 'Б': "00ab", 'В': "00ac", 'Г': "00ad",
	'Д': "00ae", 'Е': "00af", 'З': "00b1", 'И': "00b2", 'Й': "00b3",
	'К': "00b4", 'Л': "00b5", 'М': "00b6", 'Н': "00b7", 'О': "00b8",
	'П': "00b9", 'Р': "00ba", 'С': "00bb", 'Т': "00bc", 'У': "00bd",
	'Ф': "00be", 'Х': "00bf", 'Ц': "00c0", 'Ч': "00c1", 'Ш': "00c2",
	'Щ': "00c3", 'Ь': "00c6", 'Я': "00c9",
}

var chequeWideGlyphs = map[rune]string{
	' ': "0001", '₽': "00b7",
	'0': "0078", '1': "0079", '2': "007a", '3': "007b", '4': "007c",
	'5': "007d", '6': "007e", '7': "007f", '8': "0080", '9': "0081",
}

func chequeDate(value string) string {
	parsed, err := time.Parse(domain.HistoryTimeLayout, strings.TrimSpace(value))
	if err != nil {
		return strings.TrimSpace(value)
	}

	return parsed.In(time.FixedZone("MSK", 3*60*60)).Format("02.01.2006 15:04") + " ПО МСК"
}

func chequeAmount(amount float64) string {
	amount = domain.NormalizeHistoryAmount(amount)
	if math.Mod(amount, 1) == 0 {
		return strconv.FormatInt(int64(amount), 10) + " ₽"
	}

	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", amount), "0"), ".") + " ₽"
}

func chequeOperationName(item domain.HistoryItem) string {
	return strings.ToUpper(strings.TrimSpace(item.OperationFirstName + " " + item.OperationMiddleName + " " + chequeLastInitial(item.OperationLastName)))
}

func chequeClientName(clientInfo domain.ClientInfo) string {
	return strings.ToUpper(strings.TrimSpace(stringValue(clientInfo.FirstName)+" "+stringValue(clientInfo.MiddleName)+" "+chequeLastInitial(stringValue(clientInfo.LastName))) + ".")
}

func chequeLastInitial(value string) string {
	for _, char := range strings.TrimSpace(value) {
		return string(char)
	}

	return ""
}

func formatChequePhone(value string) string {
	digits := digitsOnly(value)
	if len(digits) == 11 && (digits[0] == '7' || digits[0] == '8') {
		digits = digits[1:]
	}
	if len(digits) != 10 {
		return normalizeChequeText(value)
	}

	return fmt.Sprintf("+7 %s %s-%s-%s", digits[0:3], digits[3:6], digits[6:8], digits[8:10])
}

func formatChequeCardNumber(value string) string {
	digits := digitsOnly(value)
	if digits == "" {
		return normalizeChequeText(value)
	}

	parts := make([]string, 0, (len(digits)+4)/5)
	for len(digits) > 5 {
		parts = append(parts, digits[:5])
		digits = digits[5:]
	}
	if digits != "" {
		parts = append(parts, digits)
	}

	return strings.Join(parts, " ")
}

func formatChequeMaskedCardNumber(value string) string {
	digits := digitsOnly(value)
	if len(digits) < 4 {
		return normalizeChequeText(value)
	}

	return "2200 **** **** " + digits[len(digits)-4:]
}

func cardTransferAuthCode(item domain.HistoryItem) string {
	return deterministicChequeNumber("card-auth|"+domain.HistoryItemID(item), 1000000, 6)
}

func cardTransferRRN(item domain.HistoryItem) string {
	return "613" + deterministicChequeNumber("card-rrn|"+domain.HistoryItemID(item), 1000000000, 9)
}

func deterministicChequeNumber(seed string, modulo uint64, width int) string {
	sum := sha256.Sum256([]byte(seed))
	value := binary.BigEndian.Uint64(sum[:8]) % modulo

	return fmt.Sprintf("%0*d", width, value)
}

func normalizeChequeText(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "«", `"`)
	value = strings.ReplaceAll(value, "»", `"`)

	return value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func digitsOnly(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}

	return builder.String()
}
