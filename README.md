# dev2kindle
push articles for developers to kindle

[![Build Status](https://secure.travis-ci.org/missdeer/dev2kindle.png)](https://travis-ci.org/missdeer/dev2kindle)

### Work flow
1. collect articles link from toutiao.io && iwgc.cn
2. send links to Instapaper
3. ask Instapaper to push articles to kindle
4. remove all links in Instapaper
5. wait for some minutes and loop to step 1

### Build
`go get github.com/missdeer/dev2kindle`

### Usage
`./dev2kindle -kindle xxx@kindle.com -username xxx@zzz.com -password xxxyyyzzz`

### TODO
- collect articles from gold.xitu.io && segmentfault.com/news && geek.csdn.net && gank.io
- support save links to Pocket service
- use own readability implementation
- generate .mobi file and send it to kindle mailbox by dev2kindle
