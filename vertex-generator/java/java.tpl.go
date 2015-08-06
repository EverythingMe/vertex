package java

var tpl = `
//////////////////////////////////////////////////////////////////////////////////////
//
//                  * * * CAUTION: HERE BE GENEREATED DRAGONS * * *
//
//           THIS FILE IS AUTO-GENERATED BY VERTEX. DO NOT EDIT IT MANUALLY
//
//////////////////////////////////////////////////////////////////////////////////////

/** 
* Autogenerated Class {{.Name}}
*
* {{.Doc}}
*/
public class {{.Name}} extends BaseAPI {

    static final public String ROOT = "{{.Root}}";
  
    public static class Types {
        {{ range .Types }}
        public static class {{ .Name }}{{if .Extends}} extends {{ .Extends }} {} {{else}} {
        {{ range .Members }}\
        public {{ .Type }} {{ .Name }};
        {{ end }}
        } {{end}}
        {{ end }}
    }
    
    
    public {{.Name}}(boolean secure, String host, Decoder decoder, Client client) {
        super(secure, host, decoder, client);
    }
    
    {{ range .Methods }}
        /**
        * Method {{ .Name }}
        *
        * {{ .Doc  }}
        *{{if .Params }}\
        {{ range .Params }}
        * @param {{ .Name }} {{ .Doc }}\
        {{ end }}{{end}}
        **/
        public CompletableFuture<{{ .Returns }}> {{ .Name }}({{ renderArguments .Params }}) {
            Map<String,Object> pathParams = new HashMap<>();
            Request.ParamMap params = new Request.ParamMap();
            {{ range .Params }}{{ if eq .In "query" "body" }}
            params.set("{{.Name}}", {{.Name}});\
            {{ else if eq .In "path" }}
            pathParams.put("{{.Name}}", {{.Name}});{{end}}
            {{ end}}
            return perform(Request.Method.{{ .HttpVerb }},  ROOT, "{{.Path}}",
                           params,
                           pathParams,
                           parser({{ .Returns }}.class));
        }

    {{ end }}
}
`