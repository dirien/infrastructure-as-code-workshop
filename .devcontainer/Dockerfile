FROM  --platform=linux/amd64 pulumi/pulumi:3.130.0

# Install pulumictl and set to PATH
RUN curl -sL "https://raw.githubusercontent.com/crossplane/crossplane/master/install.sh" | sh && \
    mv crossplane /usr/local/bin && \
    crossplane --help
