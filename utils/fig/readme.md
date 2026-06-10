# Create import statements and <Fig> components for mdx0

This utility should be ran from the dir containing `index.mdx` and `/images/`

```
kmsec@penguin:~/experiments/new-blog/src/content/blog/dprk-opsec-4-admin-panel$ go run ../../../../utils/fig/create-figs.go 
import BitbucketTechWorkspace from "./images/bitbucket-tech_workspace.png";
import Deployments from "./images/deployments.png";
import HeadhunterSearchResult from "./images/headhunter-search-result.png";
import Headhunter705 from "./images/headhunter705.png";
import HelenaAccountDeacivated from "./images/helena_account_deacivated.png";
import IpCheckRoute from "./images/ip-check-route.png";
import LogServerGh from "./images/log-server-gh.png";
import MaliciousEnvVar from "./images/malicious-env-var.png";
import NormalMalware from "./images/normal-malware.png";
import RequestHistoryRedacted from "./images/request_history-redacted.png";
import WalterServerCss from "./images/walter-server-css.png";
import WalterServerDashboard from "./images/walter-server-dashboard.png";
import WalterServerFilters from "./images/walter-server-filters.png";

<Fig src={BitbucketTechWorkspace} caption="changeme!" />
<Fig src={Deployments} caption="changeme!" />
<Fig src={HeadhunterSearchResult} caption="changeme!" />
<Fig src={Headhunter705} caption="changeme!" />
<Fig src={HelenaAccountDeacivated} caption="changeme!" />
<Fig src={IpCheckRoute} caption="changeme!" />
<Fig src={LogServerGh} caption="changeme!" />
<Fig src={MaliciousEnvVar} caption="changeme!" />
<Fig src={NormalMalware} caption="changeme!" />
<Fig src={RequestHistoryRedacted} caption="changeme!" />
<Fig src={WalterServerCss} caption="changeme!" />
<Fig src={WalterServerDashboard} caption="changeme!" />
<Fig src={WalterServerFilters} caption="changeme!" />
```