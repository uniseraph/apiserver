## ACS - API

### API 简述

api是用户访问acs的入口，对系统container进行过滤。

# Tree of Content

*   [Restful API](#restful-api)
    *   [Database API](#database-api)
        *   [create a database](#create-a-database)
        *   [Remove a database](#remove-a-database)
        *   [List the databases](#list-the-databases)
        *   [Inspect a database](#inspect-a-database)
        *   [set database HA](#set-database-ha)
*   [Drivers](#drivers)
    *   [Acs Database](#acsdb)

# Restful API

## Database API

### create a databse

    POST /database/create

create a database

**Example request:**

    POST /database/create HTTP/1.1
    Content-Type: application/json
    
    {
        "Name": "mydb",
        "Driver": "acsdb",
        "Options": {}
    }

**Example response:**

    HTTP/1.1 201 Created
    Content-Type: application/json
    
    {
        "Id": "22be93d5babb089c5aab8dbc369042fad48ff791584ca2da2100db837a1c7c30",
        "Warning": "",
    }

Status Codes:
* *201* - no error
* *400* - bad request
* *404* - driver not found
* *500* - server error

**JSON Parameters:**

* *Name* - The new database's name. this is a mandatory field
* *Driver* - Name of the database driver to use. Defaults to *acs* driver
* *Options* - Database specific options to be used by the drivers, see [Drivers](#drivers)


### remove a databse

    DELETE /database/(id)

remove a database

**Example request:**

    DELETE /database/remove/22be93d5babb089c5aab8dbc369042fad48ff791584ca2da2100db837a1c7c30 HTTP/1.1
    Content-Type: application/json

**Example response:**

    HTTP/1.1 204 No Content

Status Codes
* *204* - no error
* *404* - no such loadbalancer
* *500* - server error

### List the database 

    GET /database

**Example request:**

    GET /database HTTP/1.1

**Example response:**

    HTTP/1.1 200 OK
    Content-Type: application/json
    
    [
        {
            "Name": "mydb",
            "Id": "f2de39df4171b0dc801e8002d1d999b77256983dfc63041c0f34030aa3977566",
            "Driver": "acs",
                Options: {
                   "Image": "acs-reg.alipay.com/acs/mysql:1.0.0",
                    "Networks": ["vlan217"],
                    "CPU" : 2,
                    "Memory" : 1024,
                    "Disk" : 20,
            }
        }
    ]

### Inspect a database

    GET /database/<id>

**Example request:**

    GET /database/f2de39df4171b0dc801e8002d1d999b77256983dfc63041c0f34030aa3977566 HTTP/1.1

**Example response:**

    HTTP/1.1 200 OK
    Content-Type: application/json
    
    {
        "Name": "mydb",
        "Id": "f2de39df4171b0dc801e8002d1d999b77256983dfc63041c0f34030aa3977566",
        "Driver": "acs",
        Options: {
            "Image": "acs-reg.alipay.com/acs/mysql:1.0.0",
            "Networks": ["vlan217"],
            "CPU" : 2,
            "Memory" : 1024,
            "Disk" : 20,
        },
        "Attribute" : {
            "IntanceTYpe" : "Primary",    
            "Ipv4Address" : "120.69.192.13",
            "Password"    : "sFrdx9nv",
        }
    }

Status Codes:
* *200* - no error
* *404* - load balancer not found


### set database HA

    POST /database/setha

**Example request:**

    POST /database/setha HTTP/1.1
    Content-Type: application/json
    
    {
        "MasterID": "39b69226f9d79f5634485fb236a23b2fe4e96a0a94128390a7fbbcc167065867",
        "SlaveID": "f2de39df4171b0dc801e8002d1d999b77256983dfc63041c0f34030aa3977566",
        "Options": {}
    }

**Example response:**

    HTTP/1.1 204 No Content

Status Codes:
* *200* - no error
* *404* - load balancer not found


# Drivers

## acsdb

    Driver Name: acsdb

**Specific options:**

    Options: {
        "Image": "acs-reg.alipay.com/acs/mysql:1.0.0",
        "Networks": ["vlan217"],
        "CPU" : 2,
        "Memory" : 1024,
        "Disk" : 20,
    }

**JSON Parameters:**

* *Image* - The Image of acsdb
* *Networks* - The docker networks which database container should connect to
* *CPU* - The cpu of database container
* *Memory* - The memory size of database container
* *Disk* - The disk size of database, in GB

**Environment of acsdb:**

    labels:
        - com.alipay.acs.labels.system=true

