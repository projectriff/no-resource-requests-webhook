module github.com/projectriff/no-resource-requests-webhook

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	gomodules.xyz/jsonpatch/v2 v2.0.1
	// equivelent of kubernetes-1.16.3 tag for each k8s.io repo
	k8s.io/api v0.0.0-20191114100352-16d7abae0d2a
	k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/client-go v0.0.0-20191114101535-6c5935290e33
	sigs.k8s.io/controller-runtime v0.4.0
)
