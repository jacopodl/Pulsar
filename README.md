# Pulsar
Pulsar is a tool for data exfiltration and covert communication that enable you to create a secure data transfer, 
a bizzare chat or a network tunnel through different protocols, for example you can receive data from tcp connection 
and resend that to real destination through DNS packets :tada:

# Setting up Pulsar :hammer:

First, getting the code from repository:

    $ git clone https://github.com/jacopodl/pulsar
    
:warning: **Make sure you have at least Go 1.8 in your system to build Pulsar** :warning:

Compile it with followings command:

    $ cd pulsar
    $ export GOPATH=$(shell pwd)
    $ go get golang.org/x/net/icmp
    $ go build -o bin/pulsar src/main.go

or run:
    
    $ make

# Connectors :satellite:
A connector is a simple channel to the external world, with the connector you can read and write data from different sources.
* Console:
    - Default in/out connector, read data from stdin and write to stdout
* TCP
    - Read and write data through tcp connections
    
            tcp:127.0.0.1:9000
* UDP
    - Read and write data through udp packet
    
            udp:127.0.0.1:9000
* ICMP
    - Read and write data through icmp packet
    
            icmp:127.0.0.1 (the connection port is obviously useless)
* DNS
    - Read and write data through dns packet
    
            dns:fakedomain.net@127.0.0.1:1994

You can use option --in in order to select input connector and option --out to select output connector:
        
        --in tcp:127.0.0.1:9000
        --out dns:fkdns.lol:2.3.4.5:8989

# Example
In the following example Pulsar will be used to create a secure two-way tunnel on DNS protocol, data will be read from TCP connection (simple nc client) and resend encrypted through the tunnel.

    [nc 127.0.0.1 9000] <--TCP--> [pulsar] <--DNS--> [pulsar] <--TCP--> [nc -l 127.0.0.1 -p 9900]

192.168.1.198:

    $ ./pulsar --in tcp:127.0.0.1:9000 --out dns:test.org@192.168.1.199:8989 --duplex --plain in --handlers cipher 'cipher-key=supersekretkey!!'
    $ nc 127.0.0.1 9000
    
192.168.1.199:

    $ nc -l 127.0.0.1 -p 9900
    $ ./pulsar --in dns:test.org@192.168.1.199:8989 --out tcp:127.0.0.1:9900 --duplex --decode --plain out --handlers cipher 'cipher-key=supersekretkey!!'

