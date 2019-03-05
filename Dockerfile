FROM scratch

ADD kubernetes-sidecar-injector /kubernetes-sidecar-injector
    
ENTRYPOINT ["/kubernetes-sidecar-injector"]
