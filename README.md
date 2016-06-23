[![Build Status](https://travis-ci.org/namtzigla/gosw.svg?branch=master)](https://travis-ci.org/namtzigla/gosw)
# GOSW 

## Usage
Is a simple tool that allow user to switch env variables based on configuration sections.

For example for if you are working on two AWS regions you need to create a config file `~/.settings.json` that will look like:
```json
{
	"aws": {
		"_default": "lon",
		"lon": {
			"AWS_ACCESS_KEY_ID": "XXXXXXXXXX",
			"AWS_SECRET_ACCESS_KEY": "KKKKKKKKK",
			"AWS_REGION": "eu-west-1"
		},
		"syd": {
			"AWS_ACCESS_KEY_ID": "YYYYYYYYYYYYYY",
            		"AWS_SECRET_ACCESS_KEY": "ZZZZZZZZZZZZZZZZZZZ",
            		"AWS_REGION": "ap-southeast-2"	
		}
}
```

By running the command `gosw load aws lon` the command will generate the following output

```
set -ex AWS_ACCESS_KEY_ID;
set -ex AWS_SECRET_ACCESS_KEY;
set -ex AWS_REGION;
set -gx aws_name "lon";
set -gx AWS_REGION "eu-west-1";
set -gx AWS_ACCESS_KEY_ID "XXXXXXXXXX";
set -gx AWS_SECRET_ACCESS_KEY "KKKKKKKKK";
```

The output can be evaluated with by running `eval (gosw load aws lon)` 

### Complex config 
In cases where you need to run an external script/command to switch your settings the config will look like:  
```
{
    "docker": {

        "_default": "remote",
        "local": {
		"_command":"docker-machine env default"
        },
        "remote": {
            "DOCKER_HOST": "tcp://1.1.1.2:2376",
            "DOCKER_CERT_PATH": "/Users/xxx/.sdc/docker/xxx",
            "DOCKER_TLS_VERIFY":1
        }
    }
}
```

By running `gosw load docker local` the output will be:  
```
set -ex DOCKER_HOST;
set -ex DOCKER_CERT_PATH;
set -ex DOCKER_TLS_VERIFY;
set -gx docker_name "local";
eval (docker-machine env default);
```

## NOTE
This project is in a very alpha stage that generate only fish compatible scripts 
