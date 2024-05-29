package http

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/Superm4n97/html-render/pkg/template"
	"github.com/Superm4n97/html-render/pkg/template/static/assets/styles"
	gs "github.com/gorilla/schema"
	htmltemplate "html/template"
	"k8s.io/klog/v2"
	"net/http"
	"time"
)

func getTemplate(files embed.FS, t string) *htmltemplate.Template {
	return htmltemplate.Must(htmltemplate.New(t).
		Funcs(sprig.HtmlFuncMap()).
		ParseFS(files, t))
}

//var resourcesTemplate = htmltemplate.Must(htmltemplate.New("resources.gohtml").
//	Funcs(sprig.HtmlFuncMap()).
//	ParseFS(template.Files, "resources.gohtml"))

type handler struct {
	port string
}

func (h *handler) simpleIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %s!", r.URL.Path[1:])
}
func (h *handler) httpFileHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./pkg/template/index.html")
}

func mustRead(f func(name string) ([]byte, error), name string) string {
	bs, err := f(name)
	if err != nil {
		panic(err)
	}
	return string(bs)
}

type CRD struct {
	Group    string
	Resource string
	Kind     string
	Scoped   string
	Bound    bool
	Icon     string
}

type CRDsInfo struct {
	SessionID, ClusterName string
	CRDs                   []CRD
}

// See https://developers.google.com/web/fundamentals/performance/optimizing-content-efficiency/http-caching?hl=en
var noCacheHeaders = map[string]string{
	"Expires":         time.Unix(0, 0).Format(time.RFC1123),
	"Cache-Control":   "no-cache, no-store, must-revalidate, max-age=0",
	"X-Accel-Expires": "0", // https://www.nginx.com/resources/wiki/start/topics/examples/x-accel/
}

// prepareNoCache prepares headers for preventing browser caching.
func prepareNoCache(w http.ResponseWriter) {
	// Set NoCache headers
	for k, v := range noCacheHeaders {
		w.Header().Set(k, v)
	}
}

func (h *handler) handleResources(w http.ResponseWriter, r *http.Request) {
	logger := klog.FromContext(r.Context()).WithValues("method", r.Method, "url", r.URL.String())

	prepareNoCache(w)

	crdsInfo := CRDsInfo{
		SessionID:   "1234567",
		ClusterName: "my-test-cluster",
		CRDs: []CRD{
			{
				Group:    "kubedb.com",
				Resource: "mongodbs",
				Kind:     "MongoDB",
				Scoped:   "Namespaced",
				Bound:    false,
				Icon:     "https://cdn.appscode.com/k8s/icons/kubedb.com/mongodbs.svg",
			},
			{
				Group:    "kubedb.com",
				Resource: "mysqls",
				Kind:     "MySQL",
				Scoped:   "Namespaced",
				Bound:    false,
				Icon:     "https://cdn.appscode.com/k8s/icons/kubedb.com/mysqls.svg",
			},
			{
				Group:    "kubedb.com",
				Resource: "postgreses",
				Kind:     "Postgres",
				Scoped:   "Namespaced",
				Bound:    true,
				Icon:     "https://cdn.appscode.com/k8s/icons/kubedb.com/postgreses.svg",
			},
			{
				Group:    "kubedb.com",
				Resource: "kafka",
				Kind:     "Kafka",
				Scoped:   "Namespaced",
				Bound:    false,
				Icon:     "https://cdn.appscode.com/k8s/icons/kubedb.com/kafkas.svg",
			},
		},
	}
	bs := bytes.Buffer{}
	resourcesTemplate := getTemplate(template.Files, template.TemplateResources)
	if err := resourcesTemplate.Execute(&bs, crdsInfo); err != nil {
		logger.Error(err, "failed to execute template")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(bs.Bytes()) // nolint:errcheck
}

func newHandler() *handler {
	return &handler{
		port: ":8080",
	}
}
func (h *handler) addHandlers() {
	//http.HandleFunc("/", h.simpleIndexHandler)
	//http.HandleFunc("/index", h.httpFileHandler)
	//http.HandleFunc("/student", student.TemplatedHandlerBasic)
	//http.HandleFunc("/top-student", student.TemplatedHandlerFile)
	//http.HandleFunc("/crd-test", student.CrdTest)
	http.HandleFunc("/resource", h.handleResources)
	http.HandleFunc("/main.css", h.handleCSS)
	//http.HandleFunc("/bind", h.handleBind)
	//http.HandleFunc("/success", h.handleSuccess)
}

type BindForm struct {
	SessionID string   `schema:"sessionID"`
	GRs       []string `schema:"crd"`
}

var decoder = gs.NewDecoder()

func (h *handler) handleBind(w http.ResponseWriter, r *http.Request) {
	//logger := klog.FromContext(r.Context()).WithValues("method", r.Method, "url", r.URL.String())

	klog.Infof("handling bind api")
	prepareNoCache(w)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var form BindForm
	// r.PostForm is a map of our POST form values

	err = decoder.Decode(&form, r.PostForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Do something with person.Name or person.Phone
	fmt.Fprintf(w, "FORM: %+v", form)
}

type SuccessTemp struct {
	RedirectURL string
}

func (h *handler) handleSuccess(w http.ResponseWriter, r *http.Request) {
	bs := bytes.Buffer{}
	success := SuccessTemp{
		RedirectURL: "https://db.appscode.com/appscode/opscenter-linode/ui.appscode.com/v1alpha1/sections/datastore",
	}
	st := getTemplate(template.Files, template.TemplateSuccess)
	if err := st.Execute(&bs, success); err != nil {
		klog.Errorf(err.Error())
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(bs.Bytes())
}

func StartServer() {
	klog.Infof("creating http server...")
	h := newHandler()
	h.addHandlers()
	klog.Infof("starting server in port :8080...")
	if err := http.ListenAndServe(h.port, nil); err != nil {
		klog.Errorf(err.Error())
	}
}

func (h *handler) handleCSS(w http.ResponseWriter, r *http.Request) {
	bs := bytes.Buffer{}
	resourcesTemplate := getTemplate(styles.Files, template.TemplateMainCSS)
	if err := resourcesTemplate.Execute(&bs, nil); err != nil {
		klog.Infof(err.Error())
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/css")
	w.Write(bs.Bytes()) // nolint:errcheck
}
