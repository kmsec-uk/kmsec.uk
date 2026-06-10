# Create import statements and <Fig> components for mdx

This utility should be ran from the dir containing `index.mdx` and `/images/`

```
kmsec@penguin:~/experiments/new-blog/src/content/blog/dprk-google-docs$ go run ../../../../utils/fig/create-figs.go 
import ComparingDocsUsingImage from "./images/comparing-docs-using-image.png";
import DochuntRevisionHistory from "./images/dochunt-revision-history.png";
import DochuntSearchInterviews from "./images/dochunt-search-interviews.png";
import MaldocScreenshotUrlscan from "./images/maldoc-screenshot-urlscan.png";
import UrlscanSearchImageHash from "./images/urlscan-search-image-hash.png";
import UrlscanSearchPagetitle from "./images/urlscan-search-pagetitle.png";

<Fig src={ComparingDocsUsingImage} caption="changeme!" />
<Fig src={DochuntRevisionHistory} caption="changeme!" />
<Fig src={DochuntSearchInterviews} caption="changeme!" />
<Fig src={MaldocScreenshotUrlscan} caption="changeme!" />
<Fig src={UrlscanSearchImageHash} caption="changeme!" />
<Fig src={UrlscanSearchPagetitle} caption="changeme!" />
```