package beefix

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beego/bee/cmd/commands"
	"github.com/beego/bee/cmd/commands/version"
	beeLogger "github.com/beego/bee/logger"
	"github.com/beego/bee/logger/colors"
)

var CmdFix = &commands.Command{
	UsageLine: "fix",
	Short:     "Fixes your application by making it compatible with newer versions of Beego",
	Long: `As of {{"Beego 1.6"|bold}}, there are some backward compatibility issues.

  The command 'fix' will try to solve those issues by upgrading your code base
  to be compatible  with Beego version 1.6+.
`,
}

func init() {
	CmdFix.Run = runFix
	CmdFix.PreRun = func(cmd *commands.Command, args []string) { version.ShowShortVersionBanner() }
	commands.AvailableCommands = append(commands.AvailableCommands, CmdFix)
}

func runFix(cmd *commands.Command, args []string) int {
	output := cmd.Out()

	beeLogger.Log.Info("Upgrading the application...")

	dir, err := os.Getwd()
	if err != nil {
		beeLogger.Log.Fatalf("Error while getting the current working directory: %s", err)
	}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".exe") {
			return nil
		}
		err = fixFile(path)
		fmt.Fprintf(output, colors.GreenBold("\tfix\t")+"%s\n", path)
		if err != nil {
			beeLogger.Log.Errorf("Could not fix file: %s", err)
		}
		return err
	})
	beeLogger.Log.Success("Upgrade Done!")
	return 0
}

var rules = []string{
	"beego.BConfig.AppName", "beego.BConfig.AppName",
	"beego.BConfig.RunMode", "beego.BConfig.RunMode",
	"beego.BConfig.RecoverPanic", "beego.BConfig.RecoverPanic",
	"beego.BConfig.RouterCaseSensitive", "beego.BConfig.RouterCaseSensitive",
	"beego.BConfig.ServerName", "beego.BConfig.ServerName",
	"beego.BConfig.EnableGzip", "beego.BConfig.EnableGzip",
	"beego.BConfig.EnableErrorsShow", "beego.BConfig.EnableErrorsShow",
	"beego.BConfig.CopyRequestBody", "beego.BConfig.CopyRequestBody",
	"beego.BConfig.MaxMemory", "beego.BConfig.MaxMemory",
	"beego.BConfig.Listen.Graceful", "beego.BConfig.Listen.Graceful",
	"beego.BConfig.Listen.HTTPAddr", "beego.BConfig.Listen.HTTPAddr",
	"beego.BConfig.Listen.HTTPPort", "beego.BConfig.Listen.HTTPPort",
	"beego.BConfig.Listen.ListenTCP4", "beego.BConfig.Listen.ListenTCP4",
	"beego.BConfig.Listen.EnableHTTP", "beego.BConfig.Listen.EnableHTTP",
	"beego.BConfig.Listen.EnableHTTPS", "beego.BConfig.Listen.EnableHTTPS",
	"beego.BConfig.Listen.HTTPSAddr", "beego.BConfig.Listen.HTTPSAddr",
	"beego.BConfig.Listen.HTTPSPort", "beego.BConfig.Listen.HTTPSPort",
	"beego.BConfig.Listen.HTTPSCertFile", "beego.BConfig.Listen.HTTPSCertFile",
	"beego.BConfig.Listen.HTTPSKeyFile", "beego.BConfig.Listen.HTTPSKeyFile",
	"beego.BConfig.Listen.EnableAdmin", "beego.BConfig.Listen.EnableAdmin",
	"beego.BConfig.Listen.AdminAddr", "beego.BConfig.Listen.AdminAddr",
	"beego.BConfig.Listen.AdminPort", "beego.BConfig.Listen.AdminPort",
	"beego.BConfig.Listen.EnableFcgi", "beego.BConfig.Listen.EnableFcgi",
	"beego.BConfig.Listen.ServerTimeOut", "beego.BConfig.Listen.ServerTimeOut",
	"beego.BConfig.WebConfig.AutoRender", "beego.BConfig.WebConfig.AutoRender",
	"beego.BConfig.WebConfig.ViewsPath", "beego.BConfig.WebConfig.ViewsPath",
	"beego.BConfig.WebConfig.StaticDir", "beego.BConfig.WebConfig.StaticDir",
	"beego.BConfig.WebConfig.StaticExtensionsToGzip", "beego.BConfig.WebConfig.StaticExtensionsToGzip",
	"beego.BConfig.WebConfig.DirectoryIndex", "beego.BConfig.WebConfig.DirectoryIndex",
	"beego.BConfig.WebConfig.FlashName", "beego.BConfig.WebConfig.FlashName",
	"beego.BConfig.WebConfig.FlashSeparator", "beego.BConfig.WebConfig.FlashSeparator",
	"beego.BConfig.WebConfig.EnableDocs", "beego.BConfig.WebConfig.EnableDocs",
	"beego.BConfig.WebConfig.XSRFKey", "beego.BConfig.WebConfig.XSRFKey",
	"beego.BConfig.WebConfig.EnableXSRF", "beego.BConfig.WebConfig.EnableXSRF",
	"beego.BConfig.WebConfig.XSRFExpire", "beego.BConfig.WebConfig.XSRFExpire",
	"beego.BConfig.WebConfig.TemplateLeft", "beego.BConfig.WebConfig.TemplateLeft",
	"beego.BConfig.WebConfig.TemplateRight", "beego.BConfig.WebConfig.TemplateRight",
	"beego.BConfig.WebConfig.Session.SessionOn", "beego.BConfig.WebConfig.Session.SessionOn",
	"beego.BConfig.WebConfig.Session.SessionProvider", "beego.BConfig.WebConfig.Session.SessionProvider",
	"beego.BConfig.WebConfig.Session.SessionName", "beego.BConfig.WebConfig.Session.SessionName",
	"beego.BConfig.WebConfig.Session.SessionGCMaxLifetime", "beego.BConfig.WebConfig.Session.SessionGCMaxLifetime",
	"beego.BConfig.WebConfig.Session.SessionProviderConfig", "beego.BConfig.WebConfig.Session.SessionProviderConfig",
	"beego.BConfig.WebConfig.Session.SessionCookieLifeTime", "beego.BConfig.WebConfig.Session.SessionCookieLifeTime",
	"beego.BConfig.WebConfig.Session.SessionAutoSetCookie", "beego.BConfig.WebConfig.Session.SessionAutoSetCookie",
	"beego.BConfig.WebConfig.Session.SessionDomain", "beego.BConfig.WebConfig.Session.SessionDomain",
	"Ctx.Input.CopyBody(beego.BConfig.MaxMemory", "Ctx.Input.CopyBody(beego.BConfig.MaxMemorybeego.BConfig.MaxMemory",
	".URLFor(", ".URLFor(",
	".ServeJSON(", ".ServeJSON(",
	".ServeXML(", ".ServeXML(",
	".ServeJSONP(", ".ServeJSONP(",
	".XSRFToken(", ".XSRFToken(",
	".CheckXSRFCookie(", ".CheckXSRFCookie(",
	".XSRFFormHTML(", ".XSRFFormHTML(",
	"beego.URLFor(", "beego.URLFor(",
	"beego.GlobalDocAPI", "beego.GlobalDocAPI",
	"beego.ErrorHandler", "beego.ErrorHandler",
	"Output.JSONP(", "Output.JSONP(",
	"Output.JSON(", "Output.JSON(",
	"Output.XML(", "Output.XML(",
	"Input.URI()", "Input.URI()",
	"Input.URL()", "Input.URL()",
	"Input.AcceptsHTML()", "Input.AcceptsHTML()",
	"Input.AcceptsXML()", "Input.AcceptsXML()",
	"Input.AcceptsJSON()", "Input.AcceptsJSON()",
	"Ctx.XSRFToken()", "Ctx.XSRFToken()",
	"Ctx.CheckXSRFCookie()", "Ctx.CheckXSRFCookie()",
	"session.Store", "session.Store",
	".TplName", ".TplName",
	"swagger.APIRef", "swagger.APIRef",
	"swagger.APIDeclaration", "swagger.APIDeclaration",
	"swagger.API", "swagger.API",
	"swagger.APIRef", "swagger.APIRef",
	"swagger.Information", "swagger.Information",
	"toolbox.URLMap", "toolbox.URLMap",
	"logs.Logger", "logs.Logger",
	"Input.Context.Request", "Input.Context.Request",
	"Input.Params())", "Input.Params())",
	"httplib.BeegoHTTPSettings", "httplib.BeegoHTTPSettings",
	"httplib.BeegoHTTPRequest", "httplib.BeegoHTTPRequest",
	".TLSClientConfig", ".TLSClientConfig",
	".JSONBody", ".JSONBody",
	".ToJSON", ".ToJSON",
	".ToXML", ".ToXML",
	"beego.HTML2str", "beego.HTML2str",
	"beego.AssetsCSS", "beego.AssetsCSS",
	"orm.DRSqlite", "orm.DRSqlite",
	"orm.DRPostgres", "orm.DRPostgres",
	"orm.DRMySQL", "orm.DRMySQL",
	"orm.DROracle", "orm.DROracle",
	"orm.ColAdd", "orm.ColAdd",
	"orm.ColMinus", "orm.ColMinus",
	"orm.ColMultiply", "orm.ColMultiply",
	"orm.ColExcept", "orm.ColExcept",
	"GenerateOperatorSQL", "GenerateOperatorSQL",
	"OperatorSQL", "OperatorSQL",
	"orm.DebugQueries", "orm.DebugQueries",
	"orm.CommaSpace", "orm.CommaSpace",
	".DoRequest()", ".DoRequest()",
	"validation.Error", "validation.Error",
}

func fixFile(file string) error {
	rp := strings.NewReplacer(rules...)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	fixed := rp.Replace(string(content))

	// Forword the RequestBody from the replace
	// "Input.Context.Request", "Input.Context.Request",
	fixed = strings.Replace(fixed, "Input.RequestBody", "Input.RequestBody", -1)

	// Regexp replace
	pareg := regexp.MustCompile(`(Input.Params\[")(.*)("])`)
	fixed = pareg.ReplaceAllString(fixed, "Input.Param(\"$2\")")
	pareg = regexp.MustCompile(`(Input.Data\[\")(.*)(\"\])(\s)(=)(\s)(.*)`)
	fixed = pareg.ReplaceAllString(fixed, "Input.SetData(\"$2\", $7)")
	pareg = regexp.MustCompile(`(Input.Data\[\")(.*)(\"\])`)
	fixed = pareg.ReplaceAllString(fixed, "Input.Data(\"$2\")")
	// Fix the cache object Put method
	pareg = regexp.MustCompile(`(\.Put\(\")(.*)(\",)(\s)(.*)(,\s*)([^\*.]*)(\))`)
	if pareg.MatchString(fixed) && strings.HasSuffix(file, ".go") {
		fixed = pareg.ReplaceAllString(fixed, ".Put(\"$2\", $5, $7*time.Second)")
		fset := token.NewFileSet() // positions are relative to fset
		f, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			panic(err)
		}
		// Print the imports from the file's AST.
		hasTimepkg := false
		for _, s := range f.Imports {
			if s.Path.Value == `"time"` {
				hasTimepkg = true
				break
			}
		}
		if !hasTimepkg {
			fixed = strings.Replace(fixed, "import (", "import (\n\t\"time\"", 1)
		}
	}
	// Replace the v.Apis in docs.go
	if strings.Contains(file, "docs.go") {
		fixed = strings.Replace(fixed, "v.Apis", "v.APIs", -1)
	}
	// Replace the config file
	if strings.HasSuffix(file, ".conf") {
		fixed = strings.Replace(fixed, "HttpCertFile", "HTTPSCertFile", -1)
		fixed = strings.Replace(fixed, "HttpKeyFile", "HTTPSKeyFile", -1)
		fixed = strings.Replace(fixed, "EnableHttpListen", "HTTPEnable", -1)
		fixed = strings.Replace(fixed, "EnableHttpTLS", "EnableHTTPS", -1)
		fixed = strings.Replace(fixed, "EnableHttpTLS", "EnableHTTPS", -1)
		fixed = strings.Replace(fixed, "BeegoServerName", "ServerName", -1)
		fixed = strings.Replace(fixed, "AdminHttpAddr", "AdminAddr", -1)
		fixed = strings.Replace(fixed, "AdminHttpPort", "AdminPort", -1)
		fixed = strings.Replace(fixed, "HttpServerTimeOut", "ServerTimeOut", -1)
	}
	err = os.Truncate(file, 0)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, []byte(fixed), 0666)
}
