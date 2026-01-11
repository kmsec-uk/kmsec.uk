---
title: Passive Takeover - uncovering (and emulating) an expensive subdomain takeover campaign
description: This post explores an often overlooked type of subdomain takeover attack I am dubbing "passive takeover."
slug: passive-takeover
tags:
  - intel
  - shodan
  - iocs
  - T1584.001
  - subdomain takeover
date: 2023-03-05T10:00:00.000Z
---
This post explores an often overlooked type of subdomain takeover attack I am dubbing "passive takeover." This kind of attack is a non-targeted domain takeover that adversaries can use to build resources.

## Spidey sense is tingling

Whilst sleuthing Shodan, I came across a page like this:

![shodan-message.png](/images/uploads/passive-takeover/shodan-message.png)

Opening up in a browser reveals:

![visit-ip.png](/images/uploads/passive-takeover/visit-ip.png)

This site uses server-side rendering so it doesn't quite make sense yet. If you go to the domain in the certificate it makes a lot more sense:

![mysterymessage.png](/images/uploads/passive-takeover/mysterymessage.png)

This immediately piqued my curiosity. For the uninitiated, a [subdomain takeover](https://developer.mozilla.org/en-US/docs/Web/Security/Subdomain_takeovers) (Mozilla docs) attack is when someone takes control of a subdomain by creating assets that hijack dangling DNS records. Nowadays, this mostly happens in the form of a CNAME or NS attack. However, in very rare instances (spoiler: like this case) you can also take over A records.

## Scope of operation

The actor owns ~700 IPs serving this subdomain takeover message. ~650 of these IPs are running in AWS Elastic IP address space. 

Their operation is easily fingerprintable with the following shodan queries:

```
http.favicon.hash:1044386855
http.title:"subdomain takeover"
http.headers_hash:-1133016004
http.html:"I like to look at the source of websites too" <-- one false positive
"X-Subdomain-Takeover: true" <-- Custom header response from the server

```
Shodan detects a corresponding domain for ~350 of these IPs. The scope of operation is quite large. The other half don't have corresponding domains in Shodan, meaning that some of the domains' A records have been cleaned (or Shodan's data is bad).

There is no link between targeted domains. The operation appears to be entirely opportunistic.

## Understanding the actor's methodology

Using the takeover example from the first screenshot, the intriguing part of takeover is that:

- The subdomain record for `melanoma.vitaccess.com` is an A record pointing to a specific IP `35.176.199.36`
- The IP is part of Amazon AWS EC2 instance IP address space.

This raises the question: how did the actor manage to take over this subdomain using EC2? IP addresses are randomly assigned on EC2 creation. My initial theory was that they brute force the process:

1. Find a domain with a dangling A record
2. Keep spinning up an EC2 instance until they hit the jackpot

This is wildly inefficient and doesn't make sense. Firstly, spamming the creation of EC2 instances until you hit a specific IP is dumb. And secondly, the affected domains are so random that it can't be targeted. A much smarter friend came up with the following theory which ultimately forms the bedrock of the passive takeover attack:

1. Spin up an EC2 instance
2. Check passive DNS (pDNS) for a current and valid dangling A record
3. If a valid A record is found: spin up the mysterious landing page, else: abort.
4. Profit?

This is a really straightforward process but it requires two things: money and money. Money to create EC2 instances, and money to pay for passive DNS data.

## Is it passive DNS?

There *are* other ways to fulfill the operation without relying on passive DNS.

1. Check for PTR record -- unlikely to work because it's often not implemented, and the existence of a PTR record on an IP predicates the domain owner being in control of the IP.
2. Wait for HTTP connections and then lookup to see if the `Host` header points to the new instance -- this method would likely require a long-running operation that might not bear results.
3. Manual check with search engines or use public intelligence feeds.

The reason why I'm convinced this actor utilises passive DNS rather than other methods is because it's quick and effective.

### Previous research on cloud instance takeover

Whilst Googling this passive DNS takeover method, I found a discussion [on Hacker News](https://news.ycombinator.com/item?id=17021061) discussing this exact technique. Matt Bryant posted a [great overview of this kind of attack](https://bishopfox.com/blog/fishing-the-aws-ip-pool-for-dangling-domains), but he never open-sourced the tool he used. Unfortunately, this was as close as I could get to explicit research on opportunistic takeover using passive DNS.

I found a [project called eipfish](https://github.com/timkoopmans/eipfish) that is scoped specifically to the Elastic IP range and uses Shodan to check for historical records. `eipfish` is great because you can have multiple Elastic IPs (EIPs) pointing to a single EC2 instance. Unfortunately Shodan's hostname data isn't as accurate as dedicated passive DNS data, but it's a great (and cheap) tool.

All the projects I could find were scoped specifically to Elastic IPs, but this is far from just an AWS problem.

## Proof-of-concept: Passive Takeover attack
I was curious just how successful a campaign like this would be so I started scripting something to automate passive takeovers.

My script uses `vt` (the VirusTotal API commandline tool) and `doctl` (the Digital Ocean API commandline tool) for passive DNS and instance creation, respectively. Whilst this fits my tooling, the concept is simple enough to port over to any scriptable cloud provider or passive DNS data provider.

<script src="https://gist.github.com/kmsec-uk/86a3e051b284d8587a6deb93f2967c70.js"></script>

Ironically, I created this script as a loop but every time I run it, I get a domain that can be taken over in the first run. Here's an example:

<script async id="asciicast-nHK86VlTyebLGE69UEA0WXylM" src="https://asciinema.org/a/nHK86VlTyebLGE69UEA0WXylM.js"></script>

An additional note should you want to test this with DigitalOcean: You may want to just reserve an IP with the API command `doctl compute reserved-ip create`. This way you won't be billed for Droplet creation.

If you want to try this and don't have a Digital Ocean account, [here's my referral link](https://m.do.co/c/6f6dff4cca70).

#### "pwning" the domain
This part I didn't even get round to scripting because it was so short and simple.
1. SSH into the VPS
2. Download caddy
3. Write the shortest Caddyfile in existence, emulating the attacker's message:
```
dangling-host.com {
        respond "subdomain takeover on {http.request.host} - hi from kmsec.uk"
}
```
4. run caddy: `caddy start`
5. ... That's it. Caddy automatically issues a certificate from LetsEncrypt and the domain now serves the messsage:

![pwned.png](/images/uploads/passive-takeover/pwned.png)

## The Riddler of subdomain takeovers
Despite putting myself in the shoes of the actor, the actor's motivation eludes me. 

I couldn't identify any malicious intent in their operation (e.g. hijacks, scripts, advertising, phishing) to suggest this is part of a malignant operation.

They aren't trying to hide their activity, in fact quite the opposite -- their operation is loud and proud. Every HTTP request to the server responds with a custom `X-Subdomain-Takeover: true` header. The HTML contains a mischievous `<!-- I like to look at the source of websites too -->` comment. But they aren't advertising who they are or claiming clout or bounty.

Taking ownership of 700 Elastic IPs/EC2 instances just to display a mysterious message is an expensive stunt. This is [roughly](https://calculator.aws/#/addService/ElasticIP) a $2000/month operation so they are clearly well-funded. Perhaps this is a research company looking to claim some publicity? If so, why not put their logo straight on the front page? Why not inform the owners of the domain in good faith?

Perhaps this is just an eccentric side-hobby of a well-funded and patient independent actor. The Riddler of subdomain takeovers. If so, they have definitely given me some interesting food for thought!

## Closing thoughts and detection ideas
This actor was entertaining to emulate, but this operation (and my personal successes with this technique) highlights a concerning risk in cloud development and losing track of dangling assets. The risks are fairly obvious but I'll spell them out:

- Reputation damage
- Phishing
- Cookie stealing
- Fraud

And attackers have increasing reason to hunt for arbitrary targets:
- Protect operational security (piggyback on someone else's domain)
- Gaining a valid certificate
- Opportunistic extortion
- lulz

Mitigation and prevention are easy but also easy to miss: 

* **Mitigate**: simply remove the A record from your DNS provider
* **Prevent**: conduct reviews of existing infrastructure, and regularly review DNS records for dangling assets
* **Detect**: track certificate transparency logs for certificates generated for your domain. Monitor your domain with an attack surface management tool.

## Indicators
Below are a list of 345 IPs and corresponding domains the mysterious actor has taken over. This list was curated from the 700 IPs they own to exclude the ec2 naming convention (ec2-1-1-1-1.compute-1.amazonaws.com) and IPs without corresponding hostname data from Shodan. You can [view the ip-to-hostname relationship in this JSON file](/samples/passive-takeover/valid_domain_ip_matches.json).

<div class="not-prose">
<pre class="indicators text-xs font-mono font-medium language-txt no-line-numbers" data-prismjs-copy="Copy indicators">
<code>
54[.]183[.]152[.]89
13[.]57[.]6[.]80
54[.]215[.]11[.]28
52[.]53[.]227[.]116
54[.]153[.]13[.]63
54[.]153[.]98[.]98
54[.]215[.]249[.]0
54[.]183[.]254[.]37
18[.]102[.]68[.]216
54[.]84[.]193[.]220
35[.]176[.]199[.]36
3[.]80[.]161[.]15
34[.]241[.]93[.]152
54[.]154[.]50[.]127
52[.]15[.]144[.]167
15[.]223[.]120[.]197
52[.]50[.]181[.]41
54[.]193[.]100[.]173
54[.]175[.]233[.]117
3[.]14[.]6[.]128
52[.]31[.]28[.]70
18[.]232[.]185[.]95
54[.]184[.]72[.]101
18[.]144[.]35[.]64
3[.]211[.]233[.]24
13[.]56[.]211[.]182
34[.]213[.]213[.]75
52[.]26[.]221[.]152
34[.]201[.]23[.]87
54[.]208[.]247[.]209
54[.]171[.]6[.]165
34[.]210[.]26[.]205
13[.]52[.]67[.]73
3[.]235[.]183[.]114
54[.]226[.]24[.]129
13[.]56[.]182[.]105
34[.]217[.]175[.]122
52[.]14[.]156[.]161
54[.]162[.]128[.]191
54[.]219[.]192[.]82
52[.]16[.]39[.]143
13[.]56[.]158[.]158
3[.]144[.]243[.]70
3[.]122[.]101[.]108
52[.]213[.]254[.]3
54[.]244[.]4[.]46
54[.]215[.]202[.]102
54[.]174[.]134[.]251
3[.]92[.]83[.]168
52[.]211[.]73[.]251
35[.]180[.]60[.]112
99[.]79[.]53[.]38
52[.]210[.]188[.]171
107[.]21[.]25[.]71
107[.]21[.]22[.]165
52[.]53[.]252[.]57
54[.]229[.]241[.]79
54[.]67[.]55[.]15
52[.]208[.]85[.]156
35[.]178[.]5[.]74
52[.]43[.]49[.]189
3[.]93[.]40[.]187
3[.]128[.]27[.]178
54[.]187[.]193[.]102
18[.]156[.]174[.]77
18[.]144[.]27[.]237
54[.]175[.]128[.]211
54[.]183[.]13[.]242
3[.]140[.]239[.]136
54[.]241[.]85[.]172
52[.]8[.]207[.]208
54[.]245[.]14[.]50
3[.]96[.]210[.]154
3[.]131[.]94[.]71
18[.]130[.]23[.]213
52[.]205[.]74[.]24
54[.]175[.]67[.]86
3[.]101[.]111[.]143
54[.]177[.]119[.]181
13[.]56[.]13[.]177
54[.]237[.]234[.]89
3[.]136[.]106[.]125
18[.]197[.]10[.]123
54[.]226[.]21[.]102
18[.]193[.]66[.]154
34[.]229[.]146[.]29
3[.]80[.]247[.]79
34[.]242[.]230[.]20
18[.]196[.]250[.]31
34[.]243[.]97[.]168
15[.]161[.]183[.]149
174[.]138[.]109[.]102
3[.]141[.]103[.]154
107[.]23[.]130[.]249
3[.]19[.]217[.]192
3[.]140[.]184[.]198
18[.]215[.]63[.]124
34[.]244[.]11[.]56
54[.]183[.]13[.]120
52[.]87[.]251[.]186
18[.]144[.]37[.]190
34[.]204[.]45[.]179
54[.]93[.]192[.]60
164[.]92[.]102[.]57
18[.]133[.]243[.]177
35[.]182[.]73[.]42
63[.]35[.]213[.]104
34[.]207[.]120[.]173
54[.]154[.]173[.]115
52[.]53[.]224[.]187
18[.]212[.]78[.]220
34[.]204[.]194[.]143
13[.]56[.]195[.]90
174[.]129[.]54[.]26
54[.]215[.]130[.]94
34[.]213[.]214[.]22
100[.]27[.]34[.]209
54[.]89[.]102[.]233
35[.]171[.]203[.]141
3[.]135[.]202[.]134
54[.]89[.]143[.]7
176[.]34[.]89[.]47
54[.]193[.]72[.]113
18[.]130[.]52[.]28
34[.]204[.]166[.]232
54[.]203[.]122[.]109
54[.]215[.]253[.]48
3[.]101[.]88[.]130
18[.]220[.]2[.]220
54[.]84[.]240[.]212
18[.]204[.]44[.]9
52[.]72[.]21[.]236
52[.]13[.]30[.]108
34[.]219[.]154[.]25
18[.]194[.]63[.]61
54[.]172[.]125[.]59
100[.]26[.]161[.]53
34[.]254[.]247[.]113
204[.]236[.]192[.]110
23[.]23[.]12[.]156
52[.]14[.]230[.]110
35[.]181[.]53[.]198
52[.]208[.]204[.]73
52[.]202[.]126[.]38
18[.]170[.]214[.]45
3[.]8[.]15[.]244
54[.]172[.]216[.]187
18[.]182[.]61[.]135
54[.]173[.]248[.]255
54[.]226[.]135[.]173
15[.]222[.]172[.]69
54[.]198[.]232[.]255
54[.]67[.]61[.]161
54[.]163[.]195[.]48
143[.]244[.]171[.]172
35[.]183[.]34[.]135
44[.]195[.]47[.]140
13[.]40[.]177[.]14
18[.]207[.]2[.]86
54[.]201[.]8[.]176
143[.]244[.]171[.]172
143[.]244[.]171[.]172
34[.]254[.]244[.]251
3[.]145[.]97[.]4
52[.]14[.]61[.]186
13[.]52[.]219[.]86
34[.]236[.]216[.]231
54[.]195[.]179[.]20
54[.]172[.]141[.]206
54[.]145[.]47[.]245
13[.]56[.]16[.]217
54[.]152[.]127[.]5
35[.]183[.]181[.]72
54[.]197[.]207[.]51
54[.]234[.]40[.]222
3[.]228[.]219[.]148
54[.]215[.]189[.]243
54[.]158[.]205[.]242
52[.]59[.]254[.]147
18[.]219[.]104[.]42
18[.]216[.]27[.]95
54[.]183[.]235[.]221
54[.]159[.]37[.]254
54[.]152[.]8[.]201
54[.]228[.]104[.]38
54[.]219[.]221[.]26
52[.]60[.]205[.]71
54[.]170[.]2[.]34
15[.]222[.]5[.]151
13[.]38[.]244[.]156
54[.]93[.]194[.]112
34[.]247[.]73[.]73
176[.]34[.]37[.]104
35[.]180[.]22[.]43
54[.]183[.]216[.]114
52[.]23[.]213[.]167
3[.]112[.]109[.]179
35[.]78[.]247[.]172
54[.]152[.]216[.]113
107[.]20[.]26[.]160
52[.]47[.]156[.]99
3[.]226[.]251[.]249
54[.]236[.]112[.]162
52[.]53[.]239[.]223
13[.]56[.]227[.]218
3[.]8[.]208[.]132
54[.]195[.]34[.]205
52[.]30[.]222[.]12
54[.]211[.]78[.]46
107[.]22[.]146[.]224
52[.]17[.]13[.]251
3[.]64[.]165[.]171
52[.]53[.]225[.]118
13[.]40[.]114[.]210
54[.]201[.]244[.]136
54[.]215[.]214[.]160
54[.]197[.]121[.]108
54[.]93[.]124[.]195
18[.]234[.]113[.]12
3[.]72[.]60[.]246
18[.]144[.]8[.]157
13[.]57[.]185[.]59
54[.]183[.]3[.]135
13[.]59[.]4[.]150
52[.]53[.]247[.]197
3[.]231[.]26[.]52
3[.]73[.]119[.]27
54[.]183[.]250[.]235
34[.]247[.]68[.]210
18[.]184[.]218[.]162
34[.]234[.]94[.]21
54[.]250[.]238[.]221
3[.]12[.]164[.]29
54[.]193[.]62[.]18
18[.]192[.]107[.]63
3[.]93[.]156[.]240
18[.]220[.]233[.]221
54[.]215[.]208[.]99
34[.]216[.]221[.]240
15[.]188[.]89[.]38
18[.]222[.]40[.]255
13[.]48[.]132[.]234
13[.]57[.]248[.]168
107[.]23[.]249[.]140
54[.]241[.]29[.]104
54[.]85[.]68[.]213
54[.]237[.]48[.]140
54[.]64[.]226[.]215
54[.]194[.]35[.]117
35[.]183[.]127[.]99
54[.]167[.]180[.]62
34[.]204[.]91[.]238
52[.]86[.]234[.]254
54[.]183[.]88[.]23
54[.]196[.]237[.]40
52[.]53[.]218[.]168
54[.]193[.]16[.]5
18[.]191[.]147[.]240
18[.]134[.]180[.]117
3[.]239[.]205[.]23
13[.]40[.]95[.]155
54[.]193[.]134[.]12
54[.]171[.]223[.]1
34[.]220[.]227[.]11
3[.]83[.]238[.]54
54[.]90[.]176[.]46
54[.]152[.]232[.]23
15[.]188[.]26[.]180
54[.]202[.]47[.]35
54[.]86[.]115[.]24
52[.]91[.]157[.]92
54[.]153[.]45[.]28
18[.]216[.]160[.]58
54[.]193[.]75[.]220
13[.]56[.]20[.]69
18[.]222[.]237[.]175
3[.]82[.]16[.]183
3[.]101[.]83[.]25
54[.]164[.]187[.]199
54[.]219[.]145[.]19
54[.]153[.]19[.]22
52[.]52[.]110[.]78
15[.]237[.]37[.]30
35[.]172[.]141[.]73
54[.]93[.]187[.]21
54[.]183[.]138[.]36
3[.]14[.]134[.]182
3[.]231[.]55[.]233
35[.]153[.]129[.]131
13[.]37[.]253[.]17
3[.]71[.]100[.]112
44[.]193[.]203[.]151
54[.]241[.]15[.]97
52[.]91[.]138[.]156
54[.]177[.]124[.]96
146[.]190[.]220[.]214
54[.]183[.]202[.]68
18[.]117[.]150[.]251
54[.]154[.]121[.]72
54[.]183[.]143[.]203
3[.]91[.]156[.]85
54[.]166[.]166[.]79
54[.]149[.]50[.]178
52[.]91[.]51[.]173
34[.]204[.]166[.]134
52[.]16[.]101[.]205
54[.]208[.]181[.]104
54[.]173[.]74[.]242
52[.]27[.]235[.]125
35[.]182[.]242[.]119
54[.]91[.]123[.]123
3[.]12[.]197[.]178
3[.]8[.]201[.]200
18[.]220[.]162[.]130
54[.]210[.]39[.]220
44[.]200[.]35[.]84
54[.]67[.]88[.]203
18[.]216[.]143[.]247
18[.]188[.]92[.]245
54[.]84[.]203[.]243
54[.]228[.]146[.]213
54[.]172[.]98[.]159
54[.]154[.]29[.]149
13[.]49[.]241[.]153
107[.]21[.]199[.]241
18[.]188[.]20[.]60
3[.]133[.]86[.]244
35[.]167[.]214[.]162
54[.]91[.]9[.]247
18[.]220[.]118[.]244
52[.]207[.]143[.]49
3[.]220[.]167[.]140
3[.]122[.]97[.]203
54[.]173[.]202[.]117
54[.]173[.]248[.]243
18[.]195[.]97[.]242
35[.]180[.]227[.]98
50[.]18[.]145[.]41
54[.]154[.]54[.]197
107[.]23[.]82[.]11
107[.]23[.]94[.]249
3[.]140[.]208[.]210
174[.]129[.]56[.]238
35[.]153[.]177[.]232
54[.]183[.]75[.]82
education[.]sugarcrmdemo[.]com
1356226205[.]sparkstreetdigital[.]net
reports[.]mobivity[.]com
donttrip[.]technologists[.]cloud
room[.]remotepc[.]com
ian-edge[.]organization[.]arterys[.]com
ftgo[.]whochange[.]com
santanarow[.]powerflex[.]com
www[.]khtf[.]xyz
dev[.]stargreetz[.]com
melanoma[.]vitaccess[.]com
pezhawsvpn[.]softether[.]net
vpn[.]phrasee[.]co
pressrelease[.]emporioarmani[.]vespa[.]com
nfo[.]adityabirlacapital[.]com
staging-admin[.]molekule[.]com
raffle[.]nationaltrust[.]org[.]uk
dev[.]flamzy[.]com
seamus[.]deck10[.]media
certified[.]dupontregistry[.]com
p-video-wowza-15[.]video[.]kerkdienstgemist[.]nl
qa07[.]inx[.]co
licitamex[.]pinarestapalpa[.]com
test[.]aws[.]icuracao[.]com
b2bsolutions-stage[.]msv[.]mckinsey[.]com
jenkins[.]sperse[.]com
ftp[.]pcexpert[.]com[.]tw
www[.]mybusybuilding[.]com
5a26a5868d56c[.]streamlock[.]net
cms-int[.]jansport[.]com
video-wowza-63[.]video[.]kerkdienstgemist[.]nl
vpn[.]business[.]machinemetrics[.]com
relay[.]demonet[.]orbs[.]com
staging[.]surveywings[.]com
session-manager1[.]anastasiadate[.]com
js[.]sellwithleaf[.]com
old-site[.]setschedule[.]com
dev-ppp[.]iccsafe[.]org
dr-webhooks[.]feedonomics[.]com
rollback[.]mfg[.]com
sandbox-app[.]moneymover[.]com
dicom-sr[.]organization[.]arterys[.]com
vpn[.]rtfkt[.]com
id[.]lend[.]mn
nexus[.]augment[.]com
testlink[.]dev[.]identiv[.]com
c[.]clouthub[.]com
portal[.]imperson[.]com
staging-photocontest[.]rentjoy[.]com
www[.]playdota[.]com
front-03[.]oneup[.]com
bastion[.]myserve[.]co
agent[.]ct[.]web[.]identity[.]ky
identity[.]ensighten[.]com
fire[.]copart[.]com
distributor[.]kamran[.]dev-serendipity[.]wadic[.]net
nat[.]xara[.]com
kjsinclair[.]com
qa-api[.]geoplace[.]co[.]uk
h4txezppb1[.]testdrive[.]ukmdemo[.]com
portal[.]pingstart[.]com
app-dev[.]carfeine[.]com
sendy[.]localstack[.]cloud
niagaranetworks[.]com
chirpstack-new[.]pycom[.]io
www[.]mlpen[.]com
sandbox[.]cloudapi[.]nielsen[.]com
demo-connections[.]qumu[.]com
push[.]subdineapis[.]com
md[.]creativelive[.]com
demo[.]activegrid[.]com
dlma[.]azuga[.]com
www[.]btoys[.]cz
removebg-backend-old[.]invideo[.]io
photon[.]transportapi[.]com
cloud-rancher[.]misfit[.]com
unagi[.]eeygcr[.]ga
awsc[.]jiamy[.]xyz
us2[.]node[.]newbull[.]org
workbench-jail[.]psiquantum[.]com
voice[.]whisper[.]sh
order[.]globein[.]com
vasilii[.]int[.]giosg[.]com
www[.]caymanchile[.]cl
apm-dev[.]orobix[.]com
cs[.]theoreminc[.]net
arm[.]kasmweb[.]com
git2[.]tickaroo[.]com
diffusion[.]mvpworkshop[.]co
clickhouse[.]analytics[.]eu[.]silktide[.]com
www[.]jhwf[.]xyz
scheduler[.]outsite[.]co
whistleblower[.]rbw[.]it
gp[.]mogo[.]ca
events[.]forbesmiddleeast[.]com
jenkins[.]datalogics[.]com
intel-prod[.]dragos[.]com
mon[.]snapcall[.]io
lab[.]ivr[.]pindrop[.]com
aws[.]syncdog[.]com
admin[.]bestcompany[.]com
stage[.]ambbrosia[.]com
menuino[.]com
apy[.]vision
comms[.]kuflink[.]com
unifi[.]resonancedev[.]ca
www-live[.]realeyesit[.]com
shindansite[.]net
nifi-stage[.]ona[.]io
proof[.]techwire[.]net
test1[.]ialottery[.]com
demo[.]fundamentals[.]digital
start[.]tryvantagepoint[.]com
tpxe[.]warnerbros[.]com
customapparel[.]michaels[.]com
jenkins[.]dat[.]com
student4-code[.]network[.]demoredhat[.]com
twistlock[.]telus[.]digital
api-dev[.]pdffiller[.]com
sleepezee[.]bobot[.]in
helpdesk[.]tankutility[.]com
vc[.]identity[.]ky
indiaevent[.]bamkosandbox[.]com
low[.]tier2[.]solutions
cs[.]siftsecurity[.]com
vpn[.]stem[.]com
test[.]adembak[.]dpp[.]openshift[.]com
traditionsofamerica[.]glpreview[.]com
covid19[.]api[.]iottech[.]cl
invite[.]ethics[.]org
la[.]digitalaloha[.]com
beta5[.]gocanvas[.]com
bitpay[.]ravencoin[.]org
team-search[.]knote[.]com
tj294[.]37927[.]online
www[.]pen[.]privakey[.]com
jump01[.]infra[.]cloudinsights[.]netapp[.]com
av[.]halliburton[.]com
gitlab[.]fastsolucoes[.]com[.]br
api2-stage[.]weav[.]io
origin-aws[.]blogs[.]voanews[.]com
b2b-dash[.]darcmatter[.]com
dev-adv-hotfix-1097[.]webgains[.]cloud
naynas[.]com
certificate-pdf-generator[.]nuiteq[.]com
chatbot[.]thecyberhelpline[.]com
console[.]jysk[.]ca
garage2[.]dinngo[.]co
certbot[.]canstockphoto[.]com
money[.]tools[.]benzinga[.]com
ssofresh[.]firstmajestic[.]com
dataretriever[.]nrgtunnel[.]com
guoyuxin19-02c52dbf702ac4f12[.]qaboot[.]net
ice[.]wildapricot[.]com
outsite[.]co
mail[.]shavacadu[.]com
mx[.]andersencorp[.]mangoapps[.]com
helpdesk[.]ember[.]ltd
focus[.]brafton[.]com
ns3[.]gavinmc[.]com
scheduler[.]outsite[.]co
outsite[.]co
iamobile2[.]halliburton[.]com
kms[.]stg-01[.]env[.]acquire[.]io
dsspoc[.]dssdemo-se[.]dataiku[.]com
localshoppa-web-04[.]localshoppa[.]com[.]au
stage[.]mailparser[.]io
wip-devs[.]adthena[.]com
ns2[.]mgmt[.]ue1[.]pre[.]aws[.]cloud[.]arity[.]com
www[.]nutrimetermobileaws[.]horlicks[.]in
docker[.]saddleback[.]com
vpn[.]provbppr[.]com
starsboss[.]xyz
graylog-prod03[.]paperspace[.]com
beta[.]sensorcloud[.]com
codealike[.]torc[.]dev
www[.]emailstats[.]chopra[.]com
images[.]listia[.]com
legacy-latam[.]heymondo[.]com
staging-gptrac[.]sixfeetup[.]com
sql[.]solarimpulse[.]com
familynetwork[.]fandango[.]com
www[.]vbcpod2[.]com
redux[.]rhizome[.]org
training[.]infomill[.]com
6161[.]wsgr[.]com
hubot-0e50129fd8e93dc59[.]ghe-test[.]com
ukpn[.]nmcaydence[.]com
shop[.]advpharmacy[.]com
dev-dashboard[.]vegaxholdings[.]com
kviz[.]tetalena[.]cz
mm[.]screach[.]com
tokyo-test[.]sbicrypto[.]com
eu-vod-aspera[.]fubo[.]tv
us-west-1[.]vpn[.]norwoodsystems[.]io
sentry[.]infer[.]com
ugv[.]neurala[.]com
blackbox[.]walee[.]pk
kibana[.]mobilewalla[.]com
www[.]cabin7systems[.]com
sanita-localhosting-dhis2[.]bluesquare[.]org
simcanary00[.]cadesport[.]com
api-checker[.]mickael-boillaud[.]tech
vmeet-test-deploy[.]11sight[.]com
apptest[.]virginmobile[.]co
ansible-pr-347-5e2ce67b[.]feature[.]webgains[.]team
secureaccess[.]unit4[.]com
physicals[.]fastmarkets[.]com
prod-i-e8485d71[.]tnl[.]flirservices[.]com
woo[.]runalloy[.]com
git[.]xtremepush[.]com
example[.]sumo[.]app
califorec[.]cbk23[.]xyz
cloud-test[.]converis[.]clarivate[.]com
prod-demo[.]sandbox[.]truera[.]com
ftp[.]crazybaby[.]com
investorapis[.]untapped-global[.]com
sp[.]softwareag[.]com
hubtester[.]mirado1[.]info
web2[.]viewar[.]com
testsite[.]spire[.]com
dose[.]vinsol[.]com
test[.]worldplumbing[.]org
unicorp[.]venedigital[.]com
ns[.]acalvio[.]com
student1-code[.]smrtmgmt12903[.]autom8r[.]us
vc-leipzig[.]de
blog[.]avena[.]io
acc[.]kwink[.]nl
fullnode2[.]coti[.]io
api[.]ecowater[.]com
demo[.]payme[.]tokyo
prod-op-auditing[.]anytimepediatrics[.]com
timeseries[.]coinalytix[.]io
www[.]mfone[.]tech[.]co[.]ke
testwp1[.]konfirmi[.]com
training-sandbox[.]veeva[.]com
sso-saml-pwl[.]duocorp-grajput-test[.]org
sftp[.]flingo[.]tv
www[.]elrincondenico[.]com
demos[.]stage[.]huntington[.]com
logs[.]quizza[.]org
sockets[.]bestcompany[.]com
kshbdn207a[.]internal[.]demdex[.]com
stat[.]yep[.]zone
njrc[.]elitestrategies[.]dev
pegasus3[.]dataminr[.]com
ad[.]macromill[.]com
awseu-ws02[.]inbenta[.]com
artie2[.]kg-stage[.]com
3[.]prod[.]joveo[.]com
tridev[.]apple[.]fieldflex[.]com
testingeip[.]craft[.]co
labs[.]indinero[.]com
jenkins[.]presonus[.]com
mg[.]vipteacn[.]xyz
marketplace[.]organization[.]arterys[.]com
mturk[.]parl[.]ai
spark[.]bymiles[.]co[.]uk
gimli[.]arcamax[.]com
chipex[.]wiro[.]agency
www[.]startfromscratch[.]co[.]in
adv-test-maria[.]webgains[.]cloud
stage[.]appetize[.]io
botein[.]bigml[.]com
dev[.]fieldnation[.]com
enterprise-demo-2[.]coveralls[.]io
demo-instance[.]aws[.]ds[.]dataiku[.]com
wwwelk[.]aerospike[.]com
bastion[.]loadtest3[.]mountain[.]siriusxm[.]com
bastion[.]audiosocket[.]com
crvtest2[.]whitehax[.]com
meet[.]fayre[.]com
www[.]staging[.]bitdatasolutions[.]com
igen[.]acegrades[.]com
validator-api[.]tezro[.]com
stag[.]votervoice[.]net
web2[.]looksai[.]com
monitor[.]grow[.]com
ah2[.]intvenlab[.]com
mail[.]presidency[.]e-somaliland[.]com
se[.]thoughtspot[.]com
www[.]moodle[.]internal[.]babbel[.]com
smsmanager[.]comocrm[.]com
radius[.]space[.]ge
forum[.]panzura[.]com
www[.]app[.]questionbang[.]com
tnet-daemon[.]singularitynet[.]io
redirectorthingy[.]emergingtech[.]charterlab[.]com
sso[.]devtr[.]es
ocserv[.]workmotion[.]com
vpn[.]quadency[.]com
wiki[.]aquabyte[.]ai
aws-us-east-1-dev[.]0[.]dblayer[.]com
tr99999[.]com
zales[.]com
dev[.]quizbooklet[.]net
dev-voip[.]tezro[.]com
apix[.]dev[.]cloudbroker[.]vodafone[.]com
smartcredit[.]topflight[.]tech
agile[.]lftechnology[.]com
jobs2[.]floatingapps[.]com
ns2[.]nunetnetworks[.]net
sandbox2[.]innovmetric[.]com
runninghit-blog[.]inkopon[.]com
fv[.]rd[.]eu[.]clara[.]net
devteste[.]pharmaviews[.]com[.]br
www[.]harvardpartners[.]net
www[.]wasmun[.]net
meet[.]lytemedical[.]com
india[.]dealersocket[.]com
gm-poc-pm[.]my-invenio[.]com
dev-ben-api[.]d[.]bark[.]com
stg[.]lily[.]fi
dashboardqa[.]credible[.]com
cpcalendars[.]hqcoworking[.]floowmer[.]com[.]br
bull-q-ssl[.]uniqueideas[.]com
status[.]narwall[.]io
reactor-1[.]fusionauth[.]io
datameshdemomfa[.]accentureanalytics[.]com
staging[.]solgari[.]com
lightuptheholidays[.]etg[.]hearst[.]com
dev[.]notrehistoire[.]ch
wkp[.]weave[.]works
admin[.]demo[.]bonbon24h[.]com[.]vn
testnet[.]gwallet[.]tech
pilotcyber[.]orgcyberrange[.]com
www[.]stithi[.]in
rtmp[.]cyberfmradio[.]com
dealer[.]keemut[.]com
payments[.]monchilla[.]com
mktg-util-dev[.]liveperson[.]com
aws-midgard1[.]thorwallet[.]org
chat[.]development[.]macrofab[.]com
platform-legacy[.]srvr[.]acclaimworks[.]com
durazno[.]crisp[.]nl
knstl-api[.]konstellation[.]tech
jenkins[.]whitewatergames[.]net
www[.]yfha[.]xyz
onyx-it[.]mycybercns[.]com
dev[.]bunchmate[.]com
sec00payment-amazon[.]x24hr[.]com
webdev01[.]camryeffect[.]toyota[.]com
ruby3[.]honeybadger[.]io
naveen-meet[.]remotepc[.]com
</code>
</pre>
</div>