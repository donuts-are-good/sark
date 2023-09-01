![ghbanner](https://github.com/donuts-are-good/sark/assets/96031819/6b45895a-f0ec-47a2-bc6c-b88a4169c2eb)
![donuts-are-good's followers](https://img.shields.io/github/followers/donuts-are-good?&color=555&style=for-the-badge&label=followers) ![donuts-are-good's stars](https://img.shields.io/github/stars/donuts-are-good?affiliations=OWNER%2CCOLLABORATOR&color=555&style=for-the-badge) ![donuts-are-good's visitors](https://komarev.com/ghpvc/?username=donuts-are-good&color=555555&style=for-the-badge&label=visitors)

# sark

sark is a daemon for monitoring the health of cluster nodes and their applications. 

It's made to be used with [AppServe](https://github.com/donuts-are-good/appserve) but is unopinionated and can be used in any cluster environment.



## usage



### configuring sark

define your nodes and the domains they host, along with their health endpoints in `apps.json`

```json
{
  "192.168.1.10": [
    {
      "domain": "example.com",
      "healthEndpoint": "/health"
    }
  ],
  "192.168.1.11": [
    {
      "domain": "anotherexample.com",
      "healthEndpoint": "/status"
    },
    {
      "domain": "yetanotherexample.com",
      "healthEndpoint": "/check"
    }
  ]
}
```

define your preferences for the variables defined in `config.json`. these are the values that control how sark interacts with your nodes.

```json
{
  "healthCheckInterval": 60,
  "appsConfigPath": "apps.json",
  "outputFilePath": "output.txt",
  "httpClientTimeout": 10
}
```


`output.txt` is where the data goes when sark has completed a survey. maybe in the future this will be a fancy web interface, but for now, you can sysadmin your heart out and set up something to email it to yourself, or expose it on an onion site or something clever.

**note about configs:** sark scans the configs about as often as it scans the hosts and apps it is watching, so when you make a change to add a new host for example, sark does not need to be restarted and can add those changes at the start of the next interval. 

## what's up?

sark determines if a host is 'up' by if the health endpoint returns an http 200 response. if an app has crashed, usually it does not serve a 200, so we can assume the request failed and we note the time of the last successful request in our output. 


## did you know

sark is named after the character by the same name from the movie tron. sark was the character employed by the master control program to oversee the games and players on the grid. in this sense sark is the service in charge of monitoring the hosts and apps on our grid, so the name is fitting. 

## license

MIT License 2023 donuts-are-good, for more info see license.md
