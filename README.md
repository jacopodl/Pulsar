# Pulsar
Pulsar is a tool for data exfiltration and covert communication that enable you to create a secure data transfer, 
a bizzare chat or a network tunnel through different protocols, for example you can receive data from tcp connection 
and resend that to real destination through DNS packets :tada:

## Example
In the following example Pulsar will be used to create a secure two-way tunnel on DNS protocol, data will be read from TCP connection (simple nc client) and resend encrypted through the tunnel.

    [nc 127.0.0.1 9000] <--TCP--> [pulsar] <--DNS--> [pulsar] <--TCP--> [nc -l 127.0.0.1 -p 9900]

192.168.1.198:

    $ ./pulsar --in tcp:127.0.0.1:9000 --out dns:test.org@192.168.1.199:8989 --duplex --plain in --handlers cipher 'cipher-key=supersekretkey!!'
    $ nc 127.0.0.1 9000
    
192.168.1.199:

    $ nc -l 127.0.0.1 -p 9900
    $ ./pulsar --in dns:test.org@192.168.1.199:8989 --out tcp:127.0.0.1:9900 --duplex --decode --plain out --handlers cipher 'cipher-key=supersekretkey!!'

