package student

import (
	htmltemplate "html/template"
	"k8s.io/klog/v2"
	"net/http"
)

type student struct {
	Id   string
	Name string
}

func TemplatedHandlerBasic(w http.ResponseWriter, r *http.Request) {
	tmpl := htmltemplate.New("student template")
	tmpl, _ = tmpl.Parse("STUDENT: {{.Id}} - {{.Name}}!!")
	p := student{
		Id:   "123456",
		Name: "Rasel",
	}
	if err := tmpl.Execute(w, p); err != nil {
		klog.Errorf(err.Error())
		return
	}
}

func TemplatedHandlerFile(w http.ResponseWriter, r *http.Request) {
	tmpl := htmltemplate.New("top-student.html")
	var err error
	tmpl, err = tmpl.ParseFiles("./pkg/template/student.html")
	if err != nil {
		klog.Errorf(err.Error())
		return
	}

	st := student{
		Id:   "78909",
		Name: "Jamal",
	}

	if err = tmpl.Execute(w, st); err != nil {
		klog.Errorf(err.Error())
		return
	}
}

type GroupVersionKind struct {
	Group, Version, Kind string
}
type CRD struct {
	GVK    GroupVersionKind `json:"gvk"`
	Scoped string           `json:"scoped"`
	Bound  bool             `json:"bound"`
}
type CRDsInfo struct {
	SessionID   string `json:"sid"`
	ClusterName string `json:"clusterName"`
	CRDs        []CRD  `json:"crds"`
}

func CrdTest(w http.ResponseWriter, r *http.Request) {
	info := CRDsInfo{
		SessionID:   "1234567",
		ClusterName: "test-provider-cluster",
		CRDs: []CRD{
			{
				GVK: GroupVersionKind{
					Group:   "core",
					Version: "v1",
					Kind:    "pods",
				},
				Scoped: "Namespaced",
				Bound:  false,
			},
		},
	}

	tmpl := htmltemplate.New("resource.html")
	var err error
	tmpl, err = tmpl.ParseFiles("./pkg/template/resource.html")
	if err != nil {
		klog.Errorf(err.Error())
		return
	}

	if err = tmpl.Execute(w, info); err != nil {
		klog.Errorf(err.Error())
		return
	}
}
