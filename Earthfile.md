```mermaid
graph RL
    lint-golang --> test
    test-golang --> test
    test-helm --> test
    
    build-binary --> build
    build-docker --> build
    build-helm --> build
    
    build-binary --> build-docker
    release-binary --> release
    release-docker --> release
    
    build-binary --> release-binary
    build-docker --> release-docker
    
    release-helm --> continuous-deploy
    
    go-deps --> build-binary
    go-deps --> lint-golang
    go-deps --> test-golang
```
