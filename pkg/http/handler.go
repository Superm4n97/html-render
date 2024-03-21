package http

import (
	"bytes"
	"fmt"
	"github.com/Superm4n97/html-render/pkg/student"
	"github.com/Superm4n97/html-render/pkg/template"
	htmltemplate "html/template"
	"io"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"net/http"
	"time"
)

type handler struct {
	port string
}

func (h *handler) simpleIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %s!", r.URL.Path[1:])
}
func (h *handler) httpFileHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./pkg/template/index.html")
}

var resourcesTemplate = htmltemplate.Must(htmltemplate.New("resource").Parse(mustRead(template.Files.ReadFile, "resources.gohtml")))

func mustRead(f func(name string) ([]byte, error), name string) string {
	bs, err := f(name)
	if err != nil {
		panic(err)
	}
	return string(bs)
}

type CRD struct {
	GVK    schema.GroupVersionKind
	Scoped string
	Bound  bool
	Icon   string
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
				GVK: schema.GroupVersionKind{
					Group:   "catalog.appscode.com",
					Version: "v1alpha1",
					Kind:    "mysql",
				},
				Scoped: "Namespaced",
				Bound:  false,
				Icon:   "/relative/path/to/icon",
			},
			{
				GVK: schema.GroupVersionKind{
					Group:   "catalog.appscode.com",
					Version: "v1alpha1",
					Kind:    "mongodb",
				},
				Scoped: "Namespaced",
				Bound:  false,
				Icon:   "/relative/path/to/icon",
			},
			{
				GVK: schema.GroupVersionKind{
					Group:   "catalog.appscode.com",
					Version: "v1alpha1",
					Kind:    "postgresql",
				},
				Scoped: "Namespaced",
				Bound:  true,
				Icon:   "/relative/path/to/icon",
			},
		},
	}
	bs := bytes.Buffer{}
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
	http.HandleFunc("/", h.simpleIndexHandler)
	http.HandleFunc("/index", h.httpFileHandler)
	http.HandleFunc("/student", student.TemplatedHandlerBasic)
	http.HandleFunc("/top-student", student.TemplatedHandlerFile)
	http.HandleFunc("/crd-test", student.CrdTest)
	http.HandleFunc("/resource", h.handleResources)
	http.HandleFunc("/bind", h.handleBind)
}

func (h *handler) handleBind(w http.ResponseWriter, r *http.Request) {
	//logger := klog.FromContext(r.Context()).WithValues("method", r.Method, "url", r.URL.String())

	klog.Infof("handling bind api")
	prepareNoCache(w)

	data, err := io.ReadAll(r.Body)
	if err != nil {
		klog.Errorf(err.Error())
		return
	}
	klog.Infof(string(data))

	sessionID := r.URL.Query().Get("s")

	group := r.URL.Query().Get("group")
	resource := r.URL.Query().Get("resource")

	klog.Infof("session id: %s\n", sessionID)
	klog.Infof("group: %s\n", group)
	klog.Infof("resource: %s\n", resource)
}

func StartServer() {
	klog.Infof("creating http server...")
	h := newHandler()
	h.addHandlers()
	klog.Infof("starting server...")
	if err := http.ListenAndServe(h.port, nil); err != nil {
		klog.Errorf(err.Error())
	}
}
