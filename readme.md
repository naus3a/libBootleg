# libBootleg
Simple toolkit to move small amounts of data (i.e. text, light media) in a quick and secure manner across a potentially hostile environment.

Think a multiplatform, open source and cryptoanarchist AirDrop.

## Motivation
Basically I wrote this for these reasons:
* I was tired to manually copy passwords, tokens, urls across systems
* I wanted to get my feet wet with Go
* I kinda wanted to write something using [noise handshakes](http://www.noiseprotocol.org/)

Feel free to use it and/or contribute; consider it beta/active development, so think twice before using it in a life or death scenario.

## How it works
So, this is the part where we talk about [Alice & Bob](https://en.wikipedia.org/wiki/Alice_and_Bob). 
Let's say *Alice* wants to share one of those 20 characters, strong passwords with *Bob*; she could show it to him from the screen of her phone, but *Bob* would hate every second typing it. She could share a text file, but you gotta wonder what's the point of making a strong password, if you put it on a network share in plaintext. And, of course, they could use something mature and robust like `gpg` or `scp`, but *Alice* and *Bob* are lazy and don't always want to export their keys on all their systems.
They want something quick, portable, multi-platform, secure enough without feeling like it is an overkill, so they both grab a copy of the `bootlegger` tool they found in the `tools` folder of this repo.
Here's how the thing happens:
* *Bob* creates a new token:
```
bob$ bootlegger make secret 
```
* Then *Bob* shares his token with *Alice* off-channel, using a QRcode:
```
bob$ bootlegger show secret qr
```
* *Alice* scans the QRcode, extracts the token and saves it on her machine:
```
alice$ bootlegger -token bobstoken secret make
```
* The token can be  stored on both *Bob*'s and *Alice*'s system in an encrypted form, so, if they plan to share more stuff in the future, they can do so without creating new tokens every time.
* *Bob* starts listening:
```
bob$ bootlegger receive
```
* *Alice* sends her long, untypable password with:
```
alice$ send superuglypassword
```

The sender and the receiver find each other using a multicast, TOTP authenticated, discovery protocol; then they use [libdisco](https://github.com/mimoo/disco/tree/master/libdisco) to implement a [Noise NNpsk2 handshake](https://www.discocrypto.com/#/protocol/Noise_NNpsk2) and move data in a secure way.

If multicast is not an option, both sender and receiver can customize ip, port etc. Look at `bootlegger -h` to have list of what's possible.

In the real world, usually *Alice* is my phone (from where I can access my password manager etc) and *Bob* is some random system that, for one reason or another, has no access to my stuff, but needs some of it *una tantum*.
