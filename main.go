package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"bitbucket.org/pkg/inflect"
	"github.com/codegangsta/cli"
	"github.com/idada/v8.go"
)

func renderv8(o *v8.Object, format string) string {
	var s string
	if format == "html" {
		switch o.GetProperty("nodeType").ToInteger() {
		case 1:
			parts := []string{o.GetProperty("tagName").ToString()}

			attributes := o.GetProperty("attributes").ToObject()
			attributeNames := attributes.GetPropertyNames()
			if attributeNames.Length() > 0 {
				for x := 0; x < attributeNames.Length(); x++ {
					key := attributeNames.GetElement(x).ToString()
					k := attributes.GetProperty(key)
					if k.IsUndefined() || k.IsNull() {
						continue
					}
					value := attributes.GetProperty(key).ToString()
					parts = append(parts, fmt.Sprintf(`%s="%s"`, key, value))
				}
			}

			style := o.GetProperty("style").ToObject()
			styleNames := style.GetPropertyNames()
			if styleNames.Length() > 0 {
				var styles []string
				for x := 0; x < styleNames.Length(); x++ {
					key := styleNames.GetElement(x).ToString()
					k := style.GetProperty(key)
					if k.IsUndefined() || k.IsNull() {
						continue
					}
					value := style.GetProperty(key).ToString()

					switch {
					case strings.HasPrefix(key, "webkit"):
						key = "-" + inflect.Dasherize(key)
						fallthrough
					default:
						styles = append(styles, fmt.Sprintf(`%s:%s`, key, value))
					case key == "cssText":
						styles = append(styles, value)
					}
				}

				parts = append(parts, fmt.Sprintf(`style="%s"`, strings.Join(styles, ";")))
			}

			s = fmt.Sprintf("<%s>", strings.Join(parts, " "))
			children := o.GetProperty("children").ToArray()
			if children.Length() > 0 {
				for x := 0; x < children.Length(); x++ {
					s += renderv8(children.GetElement(x).ToObject(), format)
				}
			}
			s += fmt.Sprintf("</%s>", parts[0])
		case 3:
			s += o.GetProperty("data").ToString()
		}
	}
	return s
}

func run(c *cli.Context) {
	var err error
	var input []byte
	output := &bytes.Buffer{}

	if c.String("i") == "-" {
		input, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		input, err = ioutil.ReadFile(c.String("i"))
		if err != nil {
			log.Fatal(err)
		}
	}

	engine := v8.NewEngine()
	global := engine.NewObjectTemplate()
	var context *v8.Context

	// window
	global.SetAccessor("window", func(name string, info v8.AccessorCallbackInfo) {
		info.ReturnValue().Set(info.Holder().Value)
	}, nil, nil, v8.PA_None)

	// console.log()
	console_log := engine.NewFunctionTemplate(func(info v8.FunctionCallbackInfo) {
		var params []interface{}
		for x := 0; x < info.Length(); x++ {
			params = append(params, info.Get(x).ToString())
		}

		log.Println(params...)
	}, nil)

	var wg sync.WaitGroup

	// setTimeout()
	setTimeout := engine.NewFunctionTemplate(func(info v8.FunctionCallbackInfo) {
		if !info.Get(0).IsFunction() {
			log.Println("WTF")
		}

		f := info.Get(0).ToFunction()
		var d int64
		if info.Length() == 2 {
			d = info.Get(1).ToInteger()
		}
		wg.Add(1)
		time.AfterFunc(time.Millisecond*time.Duration(d), func() {
			context.Scope(func(cs v8.ContextScope) {
				defer wg.Done()
				f.Call()
			})
		})
	}, nil)
	global.SetAccessor("setTimeout", func(name string, info v8.AccessorCallbackInfo) {
		info.ReturnValue().Set(setTimeout.NewFunction())
	}, nil, nil, v8.PA_None)

	// render()
	render := engine.NewFunctionTemplate(func(info v8.FunctionCallbackInfo) {
		o := info.Get(0).ToObject()
		data := renderv8(o, info.Get(1).ToString())
		output.Write([]byte(data))
	}, nil)
	global.SetAccessor("render", func(name string, info v8.AccessorCallbackInfo) {
		info.ReturnValue().Set(render.NewFunction())
	}, nil, nil, v8.PA_None)

	// sleep()
	sleep := engine.NewFunctionTemplate(func(info v8.FunctionCallbackInfo) {
		d := time.Duration(info.Get(0).ToInteger())
		time.Sleep(d * time.Millisecond)
	}, nil)
	global.SetAccessor("sleep", func(name string, info v8.AccessorCallbackInfo) {
		info.ReturnValue().Set(sleep.NewFunction())
	}, nil, nil, v8.PA_None)

	context = engine.NewContext(global)
	context.Scope(func(cs v8.ContextScope) {
		console := engine.NewObjectTemplate()
		console.SetProperty("log", console_log.NewFunction(), v8.PA_None)
		cs.Global().SetProperty("console", console.NewObject(), v8.PA_None)
		// cs.Global().SetProperty("setTimeout", setTimeout.NewFunction(), v8.PA_None)
		// cs.Global().SetProperty("render", render.NewFunction(), v8.PA_None)

		cs.Eval(`
			function Element(tagName) {
					this.tagName = tagName;
		       this.appendChild = function(child) {
		       	this.children.push(child);
		       };

		       this.setAttribute = function(name, value) {
		         // console.log("setAttribute:", name, value);
		          this.attributes[name] = value;
		       };

		       this.attributes = {};
		       this.style = {};
		       this.children = [];
		       this.nodeType = 1; // ELEMENT_NODE
			};

			function Text(data) {
		      this.nodeType = 3; // TEXT_NODE
		      this.data = data;
			}

			var document = {};
			document.readyState = "complete";
	    document.createElement = function(tagName) {
	      return new Element(tagName);
	    };
	    document.createElementNS = function(namespaceURI, qualifiedName) {
	      return new Element(qualifiedName);
	    };
	    document.createTextNode = function(data) {
	      return new Text(data);
	    };
	    document.body = new Element("body");
	    document.implementation = {
	      hasFeature: function() { return 'SVG'; }
	    };

	    var navigator = {
	      userAgent: '',
	      vendor: ''
	    };
		`)

		rjs := RAPHAEL_JS
		if c.String("raphaeljs") != "" {
			rjs, err = ioutil.ReadFile(c.String("raphaeljs"))
			if err != nil {
				log.Fatal(err)
			}
		}
		engine.Compile(rjs, nil, nil).Run()

		engine.Compile(input, nil, nil).Run()

		cs.Eval(`
			if (document.body.children.length == 0) {
				console.log("RaphaelJS failed to render");
			} else {
			  render(document.body.children[0], "html");
			}
		`)
	})

	wg.Wait()

	if c.String("o") == "-" {
		fmt.Println(output.String())
	} else {
		fp, err := os.Create(c.String("o"))
		if err != nil {
			log.Fatal(err)
		}
		defer fp.Close()
		io.Copy(fp, output)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "raphaeljscompile"
	app.Usage = "compiles RaphaelJS to HTML without a browser"
	app.Flags = []cli.Flag{
		cli.StringFlag{"raphaeljs", "", "the raphaeljs library file (defaults to built-in 2.1.2)"},
		cli.StringFlag{"i", "-", "input file"},
		cli.StringFlag{"o", "-", "output file"},
	}
	app.Action = run
	app.Run(os.Args)
}
