# Local Website Preview

## Prerequisites

Install once:

```bash
# Hugo extended v0.128.2
curl -fsSL -o hugo_extended.deb https://github.com/gohugoio/hugo/releases/download/v0.128.2/hugo_extended_0.128.2_linux-amd64.deb
sudo dpkg -i hugo_extended.deb && rm hugo_extended.deb

# hugo-tools
curl -fsSL -o hugo-tools https://github.com/appscodelabs/hugo-tools/releases/download/v0.2.23/hugo-tools-linux-amd64
chmod +x hugo-tools && sudo mv hugo-tools /usr/local/bin/hugo-tools
```

Clone the website repo (one-time):

```bash
git clone --recurse-submodules https://github.com/appscode/kubedb.com.git ~/go/src/kubedb.dev/website
```

## Build

```bash
cd ~/go/src/kubedb.dev/website

npm install
make assets

# Point the branch entry in kubedb.json to your local branch name
hugo-tools update-branch --filename=./data/products/kubedb.json --branch=<your-branch>

# Pull all tagged doc versions from GitHub
rm -rf content/docs && mkdir -p content/docs
make docs-operator

# Overwrite the current version with your local changes
# (v2026.4.27 is the version entry that maps to your branch after update-branch)
rm -rf content/docs/v2026.4.27
cp -r ~/go/src/kubedb.dev/docs/docs content/docs/v2026.4.27
```

## Serve

```bash
hugo server --bind 0.0.0.0 --port 1313 --baseURL http://localhost:1313
```

Open http://localhost:1313

## Iterating

After editing docs, re-copy and Hugo will pick up the changes (fast render mode is on):

```bash
cp -r ~/go/src/kubedb.dev/docs/docs ~/go/src/kubedb.dev/website/content/docs/v2026.4.27
```

Or keep a watch loop running in a second terminal:

```bash
while true; do
  cp -r ~/go/src/kubedb.dev/docs/docs ~/go/src/kubedb.dev/website/content/docs/v2026.4.27
  sleep 2
done
```
