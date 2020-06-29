# Dice
#
# Step 1 prepare dice
FROM golang:alpine as dicebuilder

ENV GO111MODULE=on
ARG VERSION
ARG GIT_COMMIT
ARG BUILT

# Install git
RUN apk update && apk add git
COPY . $GOPATH/src/mypackage/myapp/
WORKDIR $GOPATH/src/mypackage/myapp/

# Get dependancies
RUN go get -d -v

# Building the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
        -ldflags="-X 'dice/utils.ServerVersion=$VERSION' -X 'dice/utils.GitCommit=$GIT_COMMIT' -X 'dice/utils.Built=$BUILT'" \
        -a -installsuffix cgo \
        -o /go/bin/dice

# Step 2 prepare probe
FROM golang:alpine as probebuilder

ENV GO111MODULE=on

# Install git
RUN apk update && apk add git
COPY ./probe/. $GOPATH/src/mypackage/myapp/
WORKDIR $GOPATH/src/mypackage/myapp/

# Get dependancies
RUN go get -d -v

# Building the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
        -a -installsuffix cgo \
        -o /go/bin/probe


# STEP 3 build release package
FROM alpine:3.11

ARG BUILT

LABEL org.label-schema.name="dice-all-in-one" \
      org.label-schema.git="https://github.com/cc4i/mahjong0.git" \
      org.label-schema.build-date=$BUILT

# For installed packages
ENV KUBE_LATEST_VERSION="v1.16.9"
ENV AWS_IAM_AUTHENTICATOR="0.5.0"
ENV HELM_VERSION="v3.2.1"
ENV GLIBC_VER="2.31-r0"

# For Dice
ENV M_WORK_HOME="/workspace"
ENV M_MODE="prod"
ENV M_S3_BUCKET_REGION="ap-southeast-1"
ENV M_S3_BUCKET="cc-mahjong-0"
ENV M_LOCAL_TILE_REPO="/workspace/tiles-repo"

# Install: bash/git/openssh/curl/nodejs/npm/kubectl/kustomize/kubeseal/awscli/jq
RUN apk add --no-cache ca-certificates \
        bash \
        git \
        openssh \
        curl \
        jq \
        nodejs \
        nodejs-npm \
    && wget -q https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -O /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl \
    && wget -q https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/v${AWS_IAM_AUTHENTICATOR}/aws-iam-authenticator_${AWS_IAM_AUTHENTICATOR}_linux_amd64 -O  /usr/local/bin/aws-iam-authenticator \
    && chmod +x /usr/local/bin/aws-iam-authenticator \
    && wget -q https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - | tar -xzO linux-amd64/helm > /usr/local/bin/helm \
    && chmod +x /usr/local/bin/helm \
    && wget https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.6.1/kustomize_v3.6.1_linux_amd64.tar.gz -O - | tar -xzO kustomize > /usr/local/bin/kustomize \
    && chmod +x /usr/local/bin/kustomize \
    && wget https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.12.4/kubeseal-linux-amd64 -O /usr/local/bin/kubeseal \
    && chmod +x /usr/local/bin/kubeseal \
    && curl -sL https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub -o /etc/apk/keys/sgerrand.rsa.pub \
    && curl -sLO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VER}/glibc-${GLIBC_VER}.apk \
    && curl -sLO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VER}/glibc-bin-${GLIBC_VER}.apk \
    && apk add --no-cache \
        glibc-${GLIBC_VER}.apk \
        glibc-bin-${GLIBC_VER}.apk \
    && curl -sL https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip -o awscliv2.zip \
    && unzip awscliv2.zip \
    && aws/install \
    && rm -rf \
        awscliv2.zip \
        aws \
        /usr/local/aws-cli/v2/*/dist/aws_completer \
        /usr/local/aws-cli/v2/*/dist/awscli/data/ac.index \
        /usr/local/aws-cli/v2/*/dist/awscli/examples \
    && npm i -g aws-cdk \
    && rm glibc-${GLIBC_VER}.apk \
    && rm glibc-bin-${GLIBC_VER}.apk \
    && rm -rf /var/cache/apk/* \
    && rm -rf /root/.config \
    && rm -rf /root/.npm \
    && echo export M_WORK_HOME=${M_WORK_HOME} >>/root/env \
    && echo export M_MODE=${M_MODE} >>/root/env \
    && echo export M_S3_BUCKET_REGION=${M_S3_BUCKET_REGION} >>/root/env \
    && echo export M_S3_BUCKET=${M_S3_BUCKET} >>/root/env \
    && echo export M_LOCAL_TILE_REPO=${M_LOCAL_TILE_REPO} >>/root/env \
    && git config --global user.email "robot@mahjong.io" \
    && git config --global user.name "robot"

# Copy dice from Step 1
COPY --from=dicebuilder /go/bin/dice /usr/local/bin/dice
# Copy probe from Step 2
COPY --from=probebuilder /go/bin/probe /usr/local/bin/probe
# Copy static files
COPY ./toy /workspace/toy
COPY ./schema /workspace/schema

WORKDIR /workspace

EXPOSE 9090
ENTRYPOINT ["/usr/local/bin/dice"]
CMD ["-c"]

# Caommands example:
#    docker run -it -v ~/mywork/mylabs/csdc/mahjong-0/tiles-repo:/workspace/tiles-repo \
#        -v ~/.aws:/root/.aws \
#        -e M_MODE=dev \
#        herochinese/dice
#
#    docker run -it -v ~/mywork/mylabs/csdc/mahjong-0/tiles-repo:/workspace/tiles-repo \
#        -v ~/.aws:/root/.aws \
#        -e M_MODE=prod
#        -e M_S3_BUCKET_REGION=ap-southeast-1
#        -e M_S3_BUCKET=cc-mahjong-0
#        herochinese/dice