# dev2kindle
push articles for developers to kindle

### Work flow
1. collect articles link from toutiao.io && gold.xitu.io && segmentfault.com/news && geek.csdn.net && gank.io && iwgc.cn
2. send links to Instapaper
3. ask Instapaper to push articles to kindle
4. remove all links in Instapaper
5. wait for some minutes and loop to step 1

### Build
`go get github.com/missdeer/dev2kindle`

### Usage
`./dev2kindle -kindle xxx@kindle.com -username xxx@zzz.com -password xxxyyyzzz`
