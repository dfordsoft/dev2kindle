# dev2kindle
push articles for developers to kindle

[![Build Status](https://secure.travis-ci.org/dfordsoft/dev2kindle.png)](https://travis-ci.org/dfordsoft/dev2kindle)

[![wercker status](https://app.wercker.com/status/31fae381bdb878554582296f8c6f14b1/m/master "wercker status")](https://app.wercker.com/project/byKey/31fae381bdb878554582296f8c6f14b1)

### Work flow
1. collect articles link from segmentfault.com/news && geek.csdn.net && gold.xitu.io && toutiao.io && iwgc.cn
2. send links to Instapaper
3. ask Instapaper to push articles to kindle
4. remove all links in Instapaper
5. wait for some minutes and loop to step 1

### Build
`go get github.com/dfordsoft/dev2kindle`

### Usage
1. modify config.json
2. run command: `./dev2kindle --config=config.json`

### TODO
- support save links to Pocket service