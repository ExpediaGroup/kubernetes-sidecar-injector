FROM scratch

ADD haystack-kube-sidecar-injector /haystack-kube-sidecar-injector
    
ENTRYPOINT ["/haystack-kube-sidecar-injector"]
