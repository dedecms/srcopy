package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dedecms/srcopy/snake"
	"github.com/jung-kurt/gofpdf"
	"github.com/urfave/cli/v2"
)

func main() {

	cli.AppHelpTemplate = `{{.Name}} {{if .Version}}{{.Version}}{{end}}

使用:
	{{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
{{if .Commands}}
命令:
	{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
选项:
	{{range .VisibleFlags}}{{.}}
	{{end}}{{end}}
`

	app := &cli.App{
		Name:    "DedeCMS Src Copy",
		Version: "v1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "title", Aliases: []string{"t"}, Usage: "以标题+目录名的方式设置PDF标题，目录名可以设置为版本号。"},
			&cli.StringFlag{Name: "path", Aliases: []string{"p"}, Usage: "源文件位置。"},
			&cli.StringFlag{Name: "files", Value: "*.php,*.htm", Aliases: []string{"f"}, Usage: "文件列表，用','进行分割，可使用通配符，如：'*.go'。"},
			&cli.BoolFlag{Name: "dir", Value: false, Aliases: []string{"d"}, Usage: "自动对当前目录下的子目录进行批处理。"},
		},
		Action: func(c *cli.Context) error {
			path := c.String("path")

			if path == "" {
				return cli.Exit("请输入需要转换的程序源文件位置，使用srcopy -h获取帮助。", 1)
			}

			exts := snake.Text(c.String("files")).Split(",")

			if c.Bool("dir") {
				for _, l := range snake.FS(path).Ls() {
					if i := snake.FS(l); i.IsDir() && strings.Index(i.Get(), ".") != 0 {
						savePDF(mergeCodes(l, exts...), c.String("title"), l)
					}
				}
			} else {
				savePDF(mergeCodes(path, exts...), c.String("title"), path)
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

// 保存Docx文件
func savePDF(src, title, out string) {
	f := snake.Text(src).ReComment()
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8Font("NotoSansSC", "", "./font/NotoSansSC-Regular.ttf")

	titleStr := title + " " + snake.FS(out).Base()
	titleStr = snake.Text(titleStr).Trim(" ").Get()

	pdf.SetTitle(titleStr, false)
	pdf.SetAuthor("Jules Verne", false)
	pdf.SetHeaderFunc(func() {
		pdf.SetFont("NotoSansSC", "", 12)
		pdf.SetLineWidth(0.1)
		pdf.Line(200, 14, 10, 14)
		pdf.Cell(0, 0, titleStr)
		pdf.Ln(10)
	})

	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetLineWidth(0.1)
		pdf.Line(200, 297-15, 10, 297-15)
		pdf.SetFont("NotoSansSC", "", 8)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(0, 10, fmt.Sprintf("第 %d 页", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetFont("NotoSansSC", "", 7)

	for _, line := range f.Lines() {
		t := snake.Text(line).Trim(" ").Trim("	").Trim("	").Trim(" ")

		ls := len(t.Get()) / 158
		if ls > 0 && len(t.Get())%158 != 0 {
			ls = ls + 1
		}

		if ls > 0 {
			spi := []int{}
			for i := 0; i < ls; i++ {
				spi = append(spi, 158*(i+1))
			}
			for _, v := range t.SplitPlace(spi) {
				pdf.Cell(0, 0, v)
				pdf.Ln(3)
			}
		} else {
			pdf.Cell(0, 0, t.Get())
			pdf.Ln(3)
		}

	}
	pdf.OutputFileAndClose(out + "/../" + titleStr + ".pdf")
}

// 根据目录，将所选择的文件合并.
func mergeCodes(path string, exts ...string) string {
	src := ""
	for _, f := range snake.FS(path).Find(exts...) {
		if t, ok := snake.FS(f).Open(); ok {
			src += t.Text().Get()
		}
	}
	return src
}
