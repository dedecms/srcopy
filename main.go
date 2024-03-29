package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/dedecms/snake"
	"github.com/jung-kurt/gofpdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
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
		Name:    "DedeCMS Src Copyright",
		Version: "v1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "软件名。"},
			&cli.StringFlag{Name: "shor", Aliases: []string{"s"}, Usage: "简称或英文名。"},
			&cli.StringFlag{Name: "ver", Usage: "软件版本号。"},
			&cli.StringFlag{Name: "sp", Value: "30", Usage: "保存前多少页"},
			&cli.StringFlag{Name: "lp", Value: "30", Usage: "保存后多少页"},
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
			var n string
			n = c.String("name") + " [简称：" + c.String("shor") + "] "

			if c.Bool("dir") {
				for _, l := range snake.FS(path).Ls() {
					n = n + snake.FS(l).Base()
					if i := snake.FS(l); i.IsDir() && strings.Index(i.Get(), ".") != 0 {

						savePDF(mergeCodes(l, exts...), n, l, c.Int("sp"), c.Int("lp"))
					}
				}
			} else {

				n = n + c.String("ver")
				savePDF(mergeCodes(path, exts...), n, path, c.Int("sp"), c.Int("lp"))
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
func savePDF(src, name, out string, sp, lp int) {
	f := snake.Text(src).ReComment()

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8Font("NotoSansSC", "", "./font/NotoSansSC-Regular.ttf")
	name = snake.Text(name).Trim(" ").Get()
	lsc := len(f.Lines())
	fmt.Println(name, "总行数：", lsc)

	pdf.SetTitle(name, false)
	pdf.SetAuthor("Jules Verne", false)
	pdf.SetHeaderFunc(func() {
		pdf.SetFont("NotoSansSC", "", 12)
		pdf.SetLineWidth(0.1)
		pdf.Line(200, 14, 10, 14)
		pdf.Cell(-1, 0, name)
		pdf.SetFont("NotoSansSC", "", 8)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(0, 0, fmt.Sprintf("第 %d 页", pdf.PageNo()), "", 0, "R", false, 0, "")
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
		l := len(t.Get())
		if l >= 150 {
			for _, v := range t.SplitInt(150) {
				pdf.Cell(0, 0, v)
				pdf.Ln(3)
			}
		} else {
			pdf.Cell(0, 0, t.Get())
			pdf.Ln(3)
		}

	}

	snake.FS(out + "/../srcopy/src/").MkDir()
	inFile := out + "/../srcopy/src/" + name + " [总行数：" + strconv.Itoa(lsc) + "].pdf"
	outFile := out + "/../srcopy/" + name + " [总行数：" + strconv.Itoa(lsc) + "].pdf"

	pdf.OutputFileAndClose(inFile)

	api.OptimizeFile(inFile, "", nil)

	if pdf.PageCount() > (lp + sp) {
		lp = pdf.PageCount() - lp
		sp = sp + 1
		api.RemovePagesFile(inFile, outFile, []string{strconv.Itoa(sp) + "-" + strconv.Itoa(lp)}, nil)
		api.OptimizeFile(outFile, "", nil)
	}

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
