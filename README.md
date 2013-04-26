Ecca Message Box
=============

The Ecca Message Box is the Secure Email platform in the Eccentric
Authentication suite.

It offers only mail boxes. They are the equivalent of
rented P.O. boxes at a post office.  Ecca mail boxes offer more or
less the same amount of privacy as physical boxes.

Each mail box receives messages. People can come over to the box and
deposit messages. Messages can be anything but it is supposed to be an
encrypted blob of bytes. Think of that as an electronic envelope so
the mail man cannot read your messages.

Like their physical counterparts, Ecca mail boxes don't do any sender
checking, spam filtering etc. It just receives the messages and stores
it to be picked up.

Every once in a while, the renter checks the box to fetch any messages
that have been deposited. The renter is the only person with the key
to open the box. It's a cryptographic private key, as we don't use
passwords anymore in the Eccentric Authentication suite.

Unlike normal P.O. boxes, there is no mail service to deliver messages
from the sender into the box. The sender has to come over to the box
and deposit it himself. But with internet connections, that's not
really a problem. It's a HTTPS-request with a POST-method. How fitting :-)


### How it works

1. The renter starts by creating an Ecca-account. It is the
cryptographic key she will use to identify herself to our post
office. With it she can create a mail box and later retrieve its
contents. The name of the account can be anything. It's only used to
identify her to us. If she wants, she genereates a long random number
for account name.

2. She logs in to our post office and asks for a post box. The post
office assigns her a box number and ties it to the account. Then it
shows her the new address. The address looks like:

    https://our.post.office/deliver/<box-number>

3. She hands out the address to every one she want to be able to send
messages to her. It's entirely up to her how she distributes the
address. If she prefers privacy, she would follow the Advanced use
section at the end.

4. People connect to our site and drop off messages destined for the
box number. The only response give to senders is whether the post
office accepts the message or not.  If the message does not get
accepted, it is proof that the intended recipient won't get it. There
can be many reasons for non-delivery. E.g. box closed, or 'full', or
not paid. Or a plain wrong box number. If the site accepts it's most
likely that the message will be delivered. However, what the receiver
does with it is up to her. Senders are wise to encrypt the message
before they hand it over.

5. Our renter wants to check the mail. She uses her ecca-account, the
one she created in step 1 to log in. We validate her certificate (the
key to the box).  She gets a page with account status and a list of
URLS to the messages. She downloads the messages and does what she
wants with them.

### API

The API is implemented as a RESTlike RPC web service. It's design is
like other Eccentric Authentication services: every action that
requires an account will need a valid client certificate. The
certificates are signed by the sites' own FPCA. It would be wise to
abolish all http-traffic, except for a redirect to https.

#### Accounts

The account API allows to create an account. It requires an
Ecca-identity from our FPCA.  The account-id is the certificate
CN. Each CN has only one account. Users are free to sign up for more
certificates, as long as the CN is sitewide unique. The FPCA takes
care of that.

An account is a convenient way to redirect the user to the FPCA to
create a certificate.

#### Open mail boxes

Users would want to open a mail box.

     GET  https://our.post.office/create-mailbox

The page shows a form with a button to press.

When pressed, the browser will do a 

     POST https://our.post.office/create-mailbox

It is required to log in with a certificate from our FPCA.  When the
certificate is valid, the system will create a new box-number. It
registers the box-number now belongs to our user who is identified by
the certificate. The server returns either a 201-created or a
4xx-error when something goes wrong.  (Eg, too many mailboxes already,
account suspended...)

Box-numbers are sitewide unique, or hell breaks lose.

With a succesful creation of a new box, the server shows the new
address. It is useful immediately.

#### To close a box, the user navigates to: 

         GET  https://our.post.office/destroy-mailbox/<box-number>

The server shows a form with a button to press.

Closing the mailbox destroys all messages still in it and makes
further drop-off attempts fail with 404-not-found.


### Drop off a message for delivery

To drop of a message a sender would call:

    POST https://our.post.office/drop-off/<box-number>

With the message in the body of the request. The message will be
stored on the server with some metadata such as time of
drop-off. Perhaps a little more, such as encoding format or media
type.  The body should be sends as is, as it is expected that the data
in it are encrypted. Any transformation of the data could invalidate
decryption, making the message unreadble.

Only the owner of the certificate that created the mailbox is supposed
to retrieve the message.

### Retrieve messages

To retrieve messages, you navigate to: 

    GET https://our.post.office/retrieve/<box-number>

And you log in with the certificate. When that validates the server
shows a page with links to the individual messages. The links look
like:

    https://our.post.office/retrieve/<box-number>/<message-id>

Your user agent can retrieve the messages by following the link. If
the sender was wise, he encrypted the message with your (recipients)
public key. Now you can decrypt and read the message. How you
interpret the message is entirely up to the receiver. The Eccentric
Authentication Message Box only does the message queueing and
delivery.

### Delete messages

To delete a message send a http DELETE to the url with the message-id.
The server is free to delete messages at all times according to its
policies. It could auto-delete after a certain time or when the
mailbox is full. 

## Advanced use

### Separate naming from addressing

The mail boxes are the delivery addresses. That's why we prescribe
that the server is creating long random numbers as
box-numbers. Although the certificate of the renter bears a CN (a user chosen
name) it would be wise not to use that certificate as the
email-name. It ties the name and address together. 

It is expected of the users that the certificate is only used to
interact with this message box service. Just to open mailboxes and
retrieve messages.

### Multiple addresses

If you sign up for an password account at a web shop, you can create a
new mail box and register that address with them. Now make a note in
your mail user agent to remember this address-shop combination. Don't
reuse the address for other shops or other people. Addresses are
cheap. Use them aplenty

Notice, messages from the shop are still in plain text, so the
operators of your Ecca Message Box can still read your messages.

Notice also that the shop can give the address to others and those
others can impersonate as the shop. There is no way for you to
recognise that a certain message is from the shop or from someone
else. There is your phishing problem.

### Encrypted messages (end of phishing)

If the shop above used Eccentric Authentication for its account
creation process - and not passwords - you would have a certificate
with them and a matching private key in your mail user agent. This
gives the shop the ability to encrypt the message with your public key
(from the certificate). Now the Message Box operators cannot read the
message anymore. Only you can with your private key. When the shop
sign their messages with their private key (part of the Ecca
protocol), you can verify that is their messages, not from someone
claiming to be them. Bye bye phishers.

If the shop sends too many messages with offers you cannot refuse,
just close the mail box. The shop will get a nice 404-not-found error
every time they keep trying.

When you are ready to buy again, log in with your ecca-account you
have at the shop and register a newly opened mail box address.

### Public known names

We can go even further, when you have an Ecca-account at, say a blog
site, you can create a signed delivery statement that tells what your
mail box address is for that identity.

You create a new mail box, sign its address with your private key
(that belongs to the public key in your certificate from the blog
site). You then publish that signed address at the blog site. Now the
bloggers (and the world) can send you messages at that address. The
signature on the signed statement also tells them what key to use to
encrypt the messages so only you can decrypt - and read - these.

A wise blogger would sign the message with delivery statement inside
the message so you know whom it was from and where to respond to.

### Improve privacy

You can improve privacy even more. All mailboxes you create with the
account you have at the Ecca Message Box are tied to the
account. Anyone who learns of the internal database of the message box
will know all the box-addressess you posess. From there it's a small
step to link those to your public signed delivery statements. Thus
linking those identities together.

To protect yourself, create multiple accounts, each with a separate
certificate. Make sure you pay with an anonymous payment system or you
can be linked by them. For even more protection, use multiple Ecca
Message Box providers.

Notice that Message Box providers will learn of the IP-addresses you
use to connect to them. Use <a href="https://torproject.org">Tor</a>
to eliminate that threat.