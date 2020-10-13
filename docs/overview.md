# Lightning Pool

Lightning Pool is a non-custodial batched uniform clearing-price auction for
Lightning Channel Lease (LCL). A LCL packages up inbound (or outbound!) channel
liquidity (ability to send/receive funds) as a fixed incoming asset (earning
interest over time) with a maturity date expressed in blocks. The maturity date
of each of the channels is enforced by Bitcoin contracts, ensuring that the
funds of the maker (the party that sold the channel) can't be swept until the
maturity height.  All cleared orders (purchased channels) are cleared in a
single batched on-chain transaction. 

The existence of an open auction to acquire/sell channel liquidity provides all
participants on the network with a more _stable_ income source in addition to
routing network fees. By selling liquidity within the marketplace, individuals
are able to price their channels to ensure that they're compensated for the
time-value of their coins within a channel, accounting for worst-case force
close CSV delays.

Pool critically allows participants on the network to exchange pricing
signals to determine where liquidity in the network is most _demanded_. A
channel opened to an area of the sub-graph that doesn't actually need that
liquidity will likely remain dormant and not earn any active routing fees.
Instead, if capital can be allocated within the network in an efficient manner,
being placed where it's most demanded, we can better utilize the allocated
capital on the network, and also allow new participants to easily identify
where their capital is most needed.

Amongst several other uses cases, the Pool allows a new participant in the
network to easily _boostrap_ their ability to receive funds by paying only a
percentage of the total amount of inbound funds acquired. As an example, a node
could acquire 100 million satoshis (1000 units, more on that below) for 100,000
satoshis, or 0.1%. Ultimately the prices will be determined by the open market
place.

A non-exhaustive list of use cases includes:

  * **Bootstrapping new users with side car channels**: A common question
    posted concerning the Lightning Network goes something like: Alice is new
    to Bitcoin entirely, how can she join the Lightning Network without her,
    herself, making any new on-chain Bitcoin transactions? It’s desirable to a
    solution to onboarding new users on to the network which is as as simple as
    sending coins to a fresh address. The Pool solves this by allowing a third
    party Carol, to purchase a channel _for_ Alice, which includes starting
    _outbound_ liquidity.

  * **Demand fueled routing node channel selection**: Another common question
    with regards to the LN is: "where should I open my channels to , such that
    they'll actually be routed through"?. Pool provides a new signal for
    autopilot agents: a market demand signal. The node can offer up its
    liquidity and have it automatically be allocated where it's most demanded.

  * **Bootstrapping new services to Lightning**: Any new service launched on
    the Lightning Network will likely need to figure out how to obtain inbound
    channels so they can accept payments. For this Pool provides an elegant
    solution in that a merchant can set up a series of "introduction points"
    negotiated via the market place. The merchant can pay a small percentage of
    the total amount of liquidity allocated towards it, and also ensure that
    the funds will be committed for a set period of time.

  * **Allowing users to instantly receive with a wallet**: A common UX
    challenge that wallets face concerns ensuring a user can receive funds as
    soon as they set up a wallet. Some wallet providers have chosen to open new
    inbound channels to users themselves. This gives users the inbound
    bandwidth they need to receive, but can come at a high capital cost to the
    wallet provider as they need to commit funds with a 1:1 ratio. The
    Lightning Pool allows them to achieve some leverage in a sense, as they can
    pay only a percentage of the funds to be allocated to a new user. As an
    eaxmple, they can pay 1000 satohis to have 1 million satoshis be alloacted
    to a user.
