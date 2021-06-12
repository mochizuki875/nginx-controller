# nginx-controller
[実践入門 Kubernetesカスタムコントローラーへの道](https://nextpublishing.jp/book/11389.html)の第5章をベースに実装したカスタムコントローラーです。

## Environment

※kubebuilderの制約上、`go version v1.15+ and < 1.16`があるためgoは1.15系をインストールした。

~~~
~ ❯❯❯ kubebuilder version 
Version: main.version{KubeBuilderVersion:"3.1.0", KubernetesVendor:"1.19.2", GitCommit:"92e0349ca7334a0a8e5e499da4fb077eb524e94a", BuildDate:"2021-05-27T17:54:28Z", GoOs:"darwin", GoArch:"amd64"}
~ ❯❯❯ go version
go version go1.15.12 darwin/amd64
~ ❯❯❯ kind version
kind v0.11.0 go1.16.4 darwin/amd64
~~~

