# Ibatiks User Guide

Ibatiks is a back office system for running a *jastip* business: buying goods
abroad on behalf of customers and shipping them once you return to Indonesia.

The application is used **only by your own team**. Customers never log in. They
keep ordering the way they always have, through WhatsApp or social media, and an
admin records those orders here.

The system exists to answer four questions that get expensive to answer by hand
once you pass a few dozen orders per trip:

1. What exactly do I need to buy on this trip, and how many of each?
2. Who paid what, and who still owes me money?
3. Which goods belong to which customer, and which are spare stock?
4. Did this trip actually make money after tickets, baggage, and shipping?

This guide is organised in menu order. Read the first two chapters once, then
use the rest as a reference for whichever screen you are on.

## Who uses it

| Role | Typical person | What they do here |
|---|---|---|
| **Owner** | Business owner | Everything, plus profit reports, store settings, and user management |
| **Admin** | Daily operator | Trips, orders, invoices, shipping, stock, receivables |
| **Tripper** | The person travelling | Opens the shopping list and records purchases from the store |

Give the tripper their own account. They will use the app on a phone while
standing in a shop in Tokyo, and they only need two screens.

## The complete cycle

Every feature in this application sits somewhere on this line. Read it once and
the rest of the manual will make sense.

```
  1  CREATE TRIP                 Dates, destination, exchange rate
       │
       ▼
  2  BUILD CATALOG               Pick products, set markup, prices locked
       │
       ▼
  3  POST TO SOCIAL MEDIA        (outside the app)
       │
       ▼
  4  RECORD ORDERS               Admin types in what customers asked for,
                                 tagging which channel each came from
       │
       ▼
  5  COLLECT DEPOSIT (DP)        Order is confirmed once money arrives
       │
       ▼
  6  SHOPPING LIST               Auto-built from every order on the trip
       │
       ▼
  7  TRIPPER BUYS                Extra units are allowed → become stock
       │
       ▼
  8  GOODS ARRIVE                Match what came back against what was ordered
       │
       ▼
  9  PACK PER CUSTOMER           One package per order; weigh, measure,
                                 and estimate the shipping cost
       │
       ▼
 10  ISSUE INVOICE               PDF + ready-to-send WhatsApp message
       │
       ▼
 11  RECEIVE FINAL PAYMENT       Order becomes fully paid
       │
       ▼
 12  SHIP VIA JNE                Enter tracking number, notify customer
       │
       ▼
 13  REPORT                      Revenue − real cost − trip expenses = profit,
                                 broken down per trip, customer, and channel
```

## Key concepts

Five ideas drive almost every rule in the system. They are worth understanding
before you use the app, because they explain why certain things are allowed and
others are refused.

**The exchange rate is locked per trip.** Every price on a trip is converted
using the one rate set when the trip was created. **Ambil kurs terkini** on the
trip form fetches today's rate from a public exchange-rate service so nobody has
to copy it out of another app — but the moment the trip is saved, that number is
frozen. If the market rate moves after the trip closes, your reported profit does
not move with it.

**Prices are snapshots, not links.** When you add a product to a trip catalog,
the selling price is calculated once and stored. When a customer orders it, that
price is copied again into the order. Editing the product master later never
changes an order that already exists.

**Cost of goods comes from real purchases.** Profit is not calculated from the
estimated cost you entered when the order was created. It comes from what the
tripper actually paid at the till.

**Surplus becomes an asset, not an expense.** If the tripper buys 8 units when
only 5 were ordered, the extra 3 go into stock. Their cost is *not* charged
against the trip's profit. The money left your pocket, but the value is still
sitting in your warehouse as goods. It becomes a cost later, when that stock
sells.

**Couriers charge for space, not just weight.** A big light box costs more to
ship than the scale suggests, because the courier bills whichever is greater:
the actual weight or the volumetric weight derived from the box dimensions. The
app calculates both, so record the dimensions when you pack.

---

# Getting Started

## Signing in

Open the application URL and enter the email and password your owner gave you.

There is no self-registration and no password reset by email. If you forget your
password, an owner resets it for you from **Pengaturan → Pengguna**.

Sessions last seven days. You stay logged in across browser restarts. Clicking
**Keluar** in the bottom-left asks for confirmation before signing you out.

## The navigation map

The sidebar is grouped by *when* you use each screen during a trip, not
alphabetically. Each group is a single row that expands in place when pressed;
opening one collapses the previous, so the list never grows back to full length.
Only **Dashboard** is a direct link, since collapsing a single item would be a
click that hides nothing.

The group holding the page you are on **expands automatically**, so arriving
anywhere shows you where you are. While collapsed, the group row prints the
active screen's name underneath its title instead.

```
  ┌──────────────────────┐
  │  Dashboard           │  ← start here every morning
  │  Perjalanan        › │
  │  Penjualan         ⌄ │  ← open: holds the current page
  │      Order           │
  │      Invoice         │
  │      Pengiriman      │
  │      Siap Kemas      │
  │  Data Master       › │
  │  Lainnya           › │
  ├──────────────────────┤
  │  Owner Ibatiks      │
  │  [ Keluar ]          │  ← sign out, asks to confirm
  └──────────────────────┘
```

| Group | Menu item | English | Used when |
|---|---|---|---|
| Ringkasan | Dashboard | Dashboard | Every morning, to see what needs attention |
| Perjalanan | Trip | Trips | Planning a trip, setting up the catalog |
| Perjalanan | Daftar Belanja | Shopping List | Standing in the shop, buying |
| Perjalanan | Pembelian | Purchases | Reviewing what was bought and where it went |
| Penjualan | Order | Orders | Recording and managing customer orders |
| Penjualan | Invoice | Invoices | Billing customers |
| Penjualan | Pengiriman | Shipments | Entering tracking numbers |
| Penjualan | Siap Kemas | Packing Queue | Working through the warehouse backlog |
| Data Master | Customer | Customers | Adding a new buyer |
| Data Master | Produk | Products | Maintaining the product catalog |
| Data Master | Stok | Stock | Managing spare goods for marketplace sale |
| Lainnya | Laporan | Reports | Chasing debts, reviewing margins, channels, customers |
| Lainnya | Pengaturan | Settings | Store identity, message templates, shipping rates, users |

## Who can see what

Menu items you have no access to are hidden, not greyed out.

| Screen | Owner | Admin | Tripper |
|---|:---:|:---:|:---:|
| Dashboard | ✓ | ✓ | ✓ |
| Trips (view) | ✓ | ✓ | ✓ |
| Trips (create, edit, catalog) | ✓ | ✓ | — |
| Shopping List | ✓ | ✓ | ✓ |
| Record purchase | ✓ | ✓ | ✓ |
| Delete purchase | ✓ | ✓ | — |
| Orders | ✓ | ✓ | — |
| Invoices | ✓ | ✓ | — |
| Shipments and packing | ✓ | ✓ | — |
| Customers | ✓ | ✓ | — |
| Products (view) | ✓ | ✓ | ✓ |
| Products (edit) | ✓ | ✓ | — |
| Stock | ✓ | ✓ | — |
| Reports: receivables, product performance, per customer, per channel | ✓ | ✓ | — |
| Reports: trip profit, order profit | ✓ | — | — |
| Shipping rates (view and estimate) | ✓ | ✓ | — |
| Shipping rates (add, delete) | ✓ | — | — |
| Settings (all tabs) | ✓ | — | — |

Profit reports are deliberately owner-only. An admin can see what customers owe,
because they need that to chase payments, but not your margins. The **Profit per
Order** tab inside Laporan is hidden entirely from admins rather than shown and
then refused.

---

# Dashboard

The landing page after login. It answers "what needs my attention today".

## The top row

Four counters that should normally be small numbers.

| Card | Indonesian label | What it counts |
|---|---|---|
| Active trips | Trip aktif | Trips that are open, closed, shopping, in transit, or arrived |
| Open orders | Order berjalan | Orders not yet completed or cancelled |
| Ready to ship | Siap dikirim | Orders fully paid but with no tracking number yet |
| Outstanding | Piutang berjalan | Total money customers still owe you |

**Ready to ship** turns amber when it is above zero. That is the queue costing
you goodwill: the customer has paid and is waiting.

## The money row

| Card | What it shows |
|---|---|
| Revenue this month | Sum of all non-cancelled, non-draft orders dated this month |
| Gross profit this month | Revenue minus the real cost of goods already purchased |
| Stock value | Quantity on hand × moving average cost, across all products |

Gross profit here excludes trip expenses. For the full picture of a single trip,
open that trip's **Profit** tab.

## The panels

**Order terbaru** lists the eight most recent orders with their status. Click an
order number to open it.

**Trip mendatang** lists trips currently open for orders, soonest departure
first.

**Produk terlaris** ranks products by units sold, with revenue, cost, and profit
per product. Negative profit here usually means a purchase price rose after you
published the selling price.

---

# Trips

A trip is one journey abroad. It owns a catalog, a set of orders, a shopping
list, expenses, and a profit report.

## Creating a trip

**Trip → Buat Trip**

| Field | Indonesian | Required | Notes |
|---|---|:---:|---|
| Title | Judul trip | Yes | Free text, e.g. "Jastip Tokyo March 2026" |
| Country | Negara | Yes | |
| City | Kota | No | |
| Departure date | Tanggal berangkat | Yes | |
| Return date | Tanggal pulang | Yes | Cannot be earlier than departure |
| Order deadline | Batas terima order | No | Leave empty for no cut-off |
| Currency | Mata uang | Yes | 3-letter code, e.g. JPY, KRW, SGD |
| Exchange rate | Kurs ke Rupiah | Yes | How many rupiah per 1 unit of the currency |
| Notes | Catatan | No | Shops to visit, baggage limits |

The exchange rate is the single most important field. Get it wrong and every
price on the trip is wrong. Use a rate slightly worse than the market rate to
give yourself a buffer.

New trips start as **Draft**. They accept no orders until you open them.

## Trip lifecycle

```
     draft ────────▶ open ◀──────▶ closed
       │              │               │
       │              └──────┬────────┘
       │                     ▼
       │                 shopping ────▶ in_transit
       │                     │               │
       │                     └───────┬───────┘
       │                             ▼
       │                          arrived
       │                             │
       │                             ▼
       │                          settled
       ▼
   cancelled  ◀── (from draft, open, closed, or shopping)
```

| Status | Indonesian | Meaning | Orders accepted? |
|---|---|---|:---:|
| `draft` | Draft | Still being set up | No |
| `open` | Buka Order | Published, taking orders | Yes |
| `closed` | Order Ditutup | Cut-off passed | No |
| `shopping` | Sedang Belanja | Tripper is buying abroad | Yes |
| `in_transit` | Perjalanan Pulang | Flying home | No |
| `arrived` | Tiba di Indonesia | Goods landed | No |
| `settled` | Selesai Dibukukan | Trip closed and booked | No |
| `cancelled` | Batal | Abandoned | No |

Orders are still accepted while the trip is `shopping`. That is deliberate. In
practice customers keep asking for one more item while the tripper is already
standing in the shop, and refusing those orders would cost you sales.

## Status changes have side effects

Every trip status button opens a confirmation box describing the effect. Two of
them move many orders at once, which is exactly why they are worth a moment's
pause. **Settled** is coloured red: there is no status after it.

Moving a trip forward drags its orders along, so you do not have to update
dozens of orders one by one.

| Trip moves to | Effect on orders |
|---|---|
| `shopping` | Orders with status `dp_paid` become `purchasing` |
| `arrived` | Orders with status `purchasing` become `arrived` |

## Trip detail: five tabs

Opening a trip gives you five tabs. They map to five different jobs.

```
  ┌──────────┬────────┬──────────┬────────┬────────┐
  │ Katalog  │ Order  │ Belanja  │ Biaya  │ Profit │
  └──────────┴────────┴──────────┴────────┴────────┘
       │         │          │         │        │
    set up    see who     buy in    record   did it
    prices    ordered     store     costs    pay off
```

---

## Tab 1 — Katalog (Catalog)

The list of products offered on this trip, each with its own cost and markup.

### How the selling price is calculated

```
   cost price in foreign currency
             │
             │  × trip exchange rate
             ▼
   cost price in rupiah  ──────────────────────┐
             │                                 │
             │  + markup                       │ this is the
             ▼                                 │ estimated
   raw selling price                           │ cost used
             │                                 │ before the
             │  round UP to nearest Rp100      │ real purchase
             ▼                                 │ is recorded
   PUBLISHED SELLING PRICE  ◀──────────────────┘
```

Two markup types are available.

| Type | Indonesian | Formula | Use when |
|---|---|---|---|
| Percent | Persen (%) | `cost × (1 + value / 100)` | Normal goods, margin scales with price |
| Nominal | Nominal (Rp) | `cost + value` | Cheap items where a percentage is too small |

### Worked examples

| Cost | Rate | Markup | Cost in Rp | Raw price | Published |
|---|---|---|---|---|---|
| ¥880 | 108.5 | 35% | Rp95,480 | Rp128,898 | **Rp128,900** |
| ¥780 | 108.5 | +Rp40,000 | Rp84,630 | Rp124,630 | **Rp124,700** |
| ¥1,000 | 100 | 30% | Rp100,000 | Rp130,000 | **Rp130,000** |

Rounding is always upward to the nearest hundred rupiah. A published price of
Rp128,898 looks careless in a social media post; Rp128,900 does not.

### Catalog fields

| Field | Indonesian | Notes |
|---|---|---|
| Product | Produk | Only products not already in this catalog appear |
| Cost price | Harga modal | In the trip currency, not rupiah |
| Markup type | Jenis markup | Percent or nominal |
| Markup value | Markup | Pre-filled from the product master, editable |
| Quota | Batas kuota | Optional cap on total units across all orders |
| Active | Aktif | Uncheck to hide from new orders without deleting |

A live preview under the form shows the resulting cost in rupiah, the selling
price, and the margin per unit before you save.

### Recalculating prices

**Hitung Ulang Harga** recomputes every catalog price using the trip's current
exchange rate.

This is a deliberate, manual action. Editing the trip's exchange rate does
**not** silently reprice a catalog you have already published to customers.
Changing prices people have already seen is a business decision, not a side
effect of fixing a typo.

### Removing a catalog item

A product that customers have already ordered cannot be removed. Uncheck
**Aktif** instead. This keeps the order history intact.

---

## Tab 2 — Order

Every order placed on this trip, with quantity, total, outstanding balance, and
status. **Catat Order** opens the order form with this trip pre-selected.

Full detail on orders is in the [Orders](#orders) chapter.

---

## Tab 3 — Belanja (Shopping)

The tripper's working screen. Covered in full in the
[Shopping List](#shopping-list) chapter, because it also has its own top-level
menu item.

---

## Tab 4 — Biaya (Expenses)

Trip costs that are not the price of goods. These are subtracted from gross
profit to give net profit.

| Category | Indonesian | Examples |
|---|---|---|
| Airfare | Tiket | Flights, seat selection |
| Baggage | Bagasi | Extra baggage allowance, overweight fees |
| Accommodation | Akomodasi | Hotel, hostel |
| Local transport | Transport | Trains, taxis, airport transfer |
| Visa | Visa | Visa fees, travel insurance |
| Other | Lainnya | SIM card, packing materials |

Each entry takes a date, description, amount, and an optional receipt URL. The
running total sits above the table.

Record these as you go. An expense you forget to enter makes the trip look more
profitable than it was, which quietly leads to underpricing the next one.

---

## Tab 5 — Profit

The financial report for the trip.

### The calculation

```
     Revenue            all non-cancelled, non-draft order totals
        −
     Cost of goods      real purchase cost allocated to those orders
   ─────────────────
   = Gross profit
        −
     Trip expenses      tickets, baggage, accommodation, transport
   ─────────────────
   = NET PROFIT
```

### What each figure means

| Figure | Indonesian | Definition |
|---|---|---|
| Revenue | Omzet | Sum of order totals, excluding cancelled and draft |
| Cost of goods | HPP barang pesanan | Real purchase cost allocated to ordered items |
| Gross profit | Laba kotor | Revenue − cost of goods |
| Trip expenses | Biaya perjalanan | Total from the Biaya tab |
| Net profit | Laba bersih | Gross profit − trip expenses |
| Margin | Margin | Net profit ÷ revenue × 100 |

### Cash flow figures

| Figure | Indonesian | Definition |
|---|---|---|
| Total capital out | Total modal keluar | **All** purchases (including surplus) + trip expenses |
| Payment received | Uang masuk dari customer | Total actually collected |
| Outstanding | Sisa tagihan belum masuk | Still owed by customers |
| Shipping billed | Ongkir ditagihkan | Shipping charged to customers |
| Shipping paid | Ongkir dibayar ke kurir | Shipping actually paid to JNE |

Note the difference between **cost of goods** and **total capital out**. Capital
out includes surplus stock; cost of goods does not.

### Why surplus stock is excluded

If the panel shows surplus stock, read it carefully. This is the single most
misunderstood number in the system.

Suppose you bought 8 units at Rp100,000 each but only 5 were ordered:

```
   8 units bought  ×  Rp100,000  =  Rp800,000 left your pocket
        │
        ├── 5 units → customer orders  →  Rp500,000 counted as COST
        │
        └── 3 units → warehouse stock  →  Rp300,000 counted as ASSET
                                           (not a cost, yet)
```

The Rp300,000 appears in **total capital out**, because the cash really did
leave. It does not appear in **cost of goods**, because you still own the goods.
It becomes a cost the day that stock sells on a marketplace.

Without this rule, a trip where you stocked up aggressively would look like a
loss even though the business is fine.

---

# Shopping List

**Daftar Belanja** — the screen the tripper uses while shopping.

Pick a trip from the dropdown at the top right. The system pre-selects the trip
that is currently open or in progress, so a tripper usually does not have to
choose anything.

## Where the list comes from

The shopping list is **not** a table someone maintains. It is calculated live
from every order on the trip, every time you open the screen.

```
   Order A (Diproses):   2 × Lotion, 1 × Snack Box
   Order B (Diproses):   3 × Lotion              ┐
   Order C (Menunggu DP): 4 × Lotion             │ grouped by product
                    │                            ┘
                    ▼
   ┌────────────────────────────────────────────────────────────┐
   │ Lotion      ordered 5   awaiting DP 4   bought 0   left 5  │
   │ Snack Box   ordered 2   awaiting DP 0   bought 0   left 2  │
   └────────────────────────────────────────────────────────────┘
```

The consequence matters in practice: if an admin in Jakarta edits an order while
the tripper is in a shop in Tokyo, the tripper's list updates on the next
refresh. There is no separate list to keep in sync.

### Only deposits count towards the buy

**Ordered** counts only orders whose deposit has been verified — status
`Diproses` and beyond. Everything still waiting for its deposit is counted
separately under **Menunggu DP** and is deliberately *not* included in what the
tripper should buy.

The reason is money, not tidiness. Buying against an order whose deposit has not
arrived means fronting the purchase with the shop's own cash, and if that
customer walks away the goods sit in stock unpaid for. The awaiting column is
still shown so the tripper can see demand building up and judge whether to chase
the deposit before leaving the shop.

A banner appears above the list whenever anything is waiting, so a shrinking
number is never mistaken for a shrinking order book.

Cancelled and draft orders are excluded entirely.

## The four summary cards

| Card | Indonesian | Meaning |
|---|---|---|
| Total ordered | Total dipesan | Units customers asked for |
| Already bought | Sudah dibeli | Units recorded as purchased |
| Remaining | Sisa belanja | Still to buy |
| Estimated capital | Perkiraan modal | Expected total cost in rupiah |

Rows where nothing remains are dimmed and marked **Lengkap** (complete).

## Recording a purchase

Press **Catat Beli** on a row.

| Field | Indonesian | Notes |
|---|---|---|
| Quantity | Jumlah dibeli | Defaults to the remaining quantity |
| Unit price | Harga satuan | In trip currency, what you actually paid |
| Purchase date | Tanggal beli | Defaults to today |
| Custom rate | Kurs khusus | Optional; leave empty to use the trip rate |
| Store | Toko | Where you bought it |
| Notes | Catatan | Variant, promo, anything worth remembering |

Enter the price you actually paid, including any discount you got. This is what
makes the profit report honest.

The **custom rate** field exists because a long trip can span several days of
moving exchange rates. Each purchase keeps the rate that applied when it
happened.

## What happens on save

The system allocates the units automatically. Orders are filled oldest first, so
the customer who ordered earliest is served first when supply is tight.

```
   Bought 8 units of Lotion
        │
        ▼
   ┌────────────────────────────────────────────────────────┐
   │  Order A (ordered 15 Aug, needs 2)   ──▶  2 units      │
   │  Order B (ordered 16 Aug, needs 3)   ──▶  3 units      │
   │                                          ─────────     │
   │                             allocated to orders: 5     │
   │                                                        │
   │  Nobody ordered the rest             ──▶  3 units      │
   │                             moved into stock:      3   │
   └────────────────────────────────────────────────────────┘
```

A confirmation message tells you the split, for example
"5 units for orders, 3 units into stock".

If you buy fewer units than were ordered, the earliest orders are filled and the
rest stay pending. Their fulfilment status becomes **Sebagian** (partial).

---

# Purchases

**Pembelian** — the audit trail of everything bought, across all trips.

Filter by trip or search by product or store name. Each row shows quantity, unit
cost in both currencies, and the total in rupiah.

Two badges show where the units went:

| Badge | Meaning |
|---|---|
| *N* pesanan | *N* units allocated to customer orders |
| *N* stok | *N* units that became warehouse stock |

Expand a row with the arrow on the left to see exactly which order and customer
received each unit.

## Deleting a purchase

Deleting reverses everything the purchase caused:

- Allocations to orders are released
- Stock that was added is pulled back out
- The purchased quantity on affected order items is recalculated

If the surplus stock has already been sold, the reversal will fail because the
stock is no longer there to remove. Correct it with a stock adjustment instead.

---

# Orders

**Order** — the heart of the system.

## Order list

Search by order number, customer name, recipient, or phone. Filter by trip,
status, sales channel, or **Belum lunas** (not fully paid).

Outstanding balances are shown in amber so unpaid orders stand out.

The **Channel** column shows where each order came in from, colour-coded so a
glance down the list tells you which channel is busiest.

### Viewing totals in the trip currency

Tick **Mata uang trip** and every money column switches from rupiah to the
currency of that order's trip, converted at the exchange rate locked when the
trip was created.

Orders from different trips convert at their own rates, so a list mixing a Japan
trip and a Korea trip shows JPY on some rows and KRW on others. Nothing is
recalculated in the database; this is a display toggle only, and orders on an
IDR trip stay in rupiah.

Use it when you are reconciling against receipts from the shop abroad, where
every figure is in the local currency.

## Creating an order

**Catat Order** opens a three-part form.

### Part 1 — Trip and customer

Pick the trip first. The product picker is empty until you do, because prices
come from that trip's catalog.

Changing the trip clears any items you already selected, since the prices would
no longer apply.

**Asal order** records which channel the order arrived through. It defaults to
WhatsApp because that is where most jastip orders come from, and it feeds the
Per Channel report.

| Value | Indonesian | Use for |
|---|---|---|
| `whatsapp` | WhatsApp | Direct chat, broadcast lists, groups |
| `instagram` | Instagram | Story replies, DMs, comments |
| `tiktok` | TikTok | Comments and DMs from videos or lives |
| `marketplace` | Marketplace | Shopee, Tokopedia, and similar |
| `lainnya` | Lainnya | Anything else: walk-ins, referrals, phone calls |

Pick it as you type the order. Reconstructing where an order came from a week
later is guesswork, and guesswork makes the channel report worthless.

It can be corrected later from **Ubah Order** if you picked the wrong one.

### Part 2 — Products

Choose from the trip catalog. Each addition brings its catalog price, which you
may override per order for a special discount.

Selecting the same product twice increases its quantity instead of creating a
duplicate line.

If a product has a quota and your quantity would exceed it, a warning appears
under the line and the server will reject the order on save.

### Part 3 — Shipping address

By default the order ships to the customer's saved address, shown for
confirmation.

Tick **Kirim ke alamat lain** to send it elsewhere: a gift, an office, a
friend's house. The customer's address is copied in as a starting point so you
only change what differs.

The address is **copied into the order**, not linked. If the customer moves next
year, this order still shows where the goods actually went.

### The summary panel

| Field | Indonesian | Notes |
|---|---|---|
| Discount | Diskon | Cannot exceed the subtotal |
| Shipping fee | Ongkir ditagihkan | What you charge the customer |
| Deposit required | DP diminta | Leave empty for the 50% default |
| Notes | Catatan | Special requests |

```
   Subtotal            sum of quantity × unit price
      − Discount
      + Shipping fee
   ─────────────────
   = TOTAL
      − Paid amount
   ─────────────────
   = BALANCE DUE
```

## Order lifecycle

A newly recorded order starts at **awaiting_dp**, not `draft` — recording it in
the back office *is* the confirmation, and the very next thing you do is ask for
the deposit. `draft` still exists for orders parked mid-entry.

```
              draft
                │
                ▼
           awaiting_dp ◀────┐  ← new orders start here
                │           │  (revert if
                ▼           │   deposit fails)
             dp_paid ───────┘
                │
                ▼
           purchasing
                │
                ▼
            arrived
                │
                ▼
             packed ◀───────┐
                │           │  (repack if
                ▼           │   invoice voided)
            invoiced ───────┘
                │
                ▼
              paid
                │
                ▼
            shipped
                │
                ▼
           completed

   cancelled  ◀── from any state up to and including invoiced
```

| Status | Indonesian | Meaning |
|---|---|---|
| `draft` | Draft | Parked mid-entry, not yet confirmed |
| `awaiting_dp` | Menunggu DP | Confirmed, waiting for the deposit — where new orders start |
| `dp_paid` | Diproses | Deposit received and verified; the order now counts towards the shopping list |
| `purchasing` | Dibelikan | Tripper is buying the goods |
| `arrived` | Barang Tiba | Goods received and matched |
| `packed` | Sudah Dikemas | Packed under the customer's name |
| `invoiced` | Ditagihkan | Final invoice sent |
| `paid` | Lunas | Fully paid, ready to ship |
| `shipped` | Dikirim | Handed to the courier, tracking recorded |
| `completed` | Selesai | Confirmed received by the customer |
| `cancelled` | Batal | Cancelled |

Every status button opens a confirmation box first, stating what the move does
rather than merely asking whether you are sure. Status changes have no undo, and
several of them lock the order, so the box is there to be read — the move to
`shipped` is coloured red for that reason.

Two transitions are guarded by money, not by a button:

- You cannot move to `dp_paid` until deposits actually recorded reach the
  required amount.
- You cannot move to `paid` while any balance is outstanding.

## Editing order items

This is the most frequently used action in daily operation. A customer messages
"can you make that two instead of one" and you change it here.

Each line can be edited in place with the pencil icon.

### The rules

| Rule | Why |
|---|---|
| Cannot edit at all once the order is `shipped`, `completed`, or `cancelled` | The package is gone; documents must match reality |
| Quantity cannot go below the quantity already received | You cannot un-receive goods that are in your warehouse |
| Reducing below the quantity already purchased releases the excess into stock | The goods physically exist; they must not vanish from the books |
| An order must keep at least one item | Cancel the order instead |
| Discount cannot exceed the subtotal | Prevents a negative total |

### What happens automatically

```
   You change quantity from 3 to 2
        │
        ├─▶ Line subtotal recalculated
        ├─▶ Order subtotal and total recalculated
        ├─▶ Balance due recalculated
        ├─▶ 1 surplus unit released from this order into stock
        ├─▶ Status stepped back if the order was paid but now owes money
        └─▶ Entry written to the audit trail
```

That fifth step matters. If an order was fully paid and you *add* items, the
status drops from `paid` back to `invoiced` so the new balance is visible
instead of hiding behind a green "paid" badge.

## Payments

The **Pembayaran** panel shows deposit required, amount paid, and balance due,
followed by every payment recorded.

| Type | Indonesian | Effect on balance |
|---|---|---|
| Deposit | DP | Increases paid amount |
| Settlement | Pelunasan | Increases paid amount |
| Refund | Refund | Decreases paid amount |
| Adjustment | Penyesuaian | Increases paid amount |

Methods available: bank transfer, cash, QRIS, e-wallet, other.

### Recording a payment

**Catat Bayar** pre-fills the amount with whatever is outstanding: the remaining
deposit if the deposit is not yet complete, otherwise the full balance.

Optional fields: reference (transfer ID or sender name), proof of transfer,
payment date, and notes.

**Bukti transfer** uploads the customer's transfer receipt — a photo or a PDF —
to the shop's own server; JPG, PNG, WEBP, and PDF are accepted. Images show a
thumbnail so you can confirm you attached the right screenshot before saving, and
the file stays openable from the payment list afterwards for reconciling against
the bank statement.

The file type is detected from the file's own contents rather than its name, and
it is stored under a fresh random filename, so a misnamed or hostile upload
cannot decide where it lands on disk.

### Automatic status changes

| Condition after payment | Order becomes |
|---|---|
| Deposit reaches the required amount, order was `awaiting_dp` | `dp_paid` |
| Balance reaches zero | `paid` |

A refund larger than the amount received is rejected.

### Requesting a deposit

While a deposit is outstanding, a **Tagih DP** button appears. It opens a
ready-to-send WhatsApp message built from your template, with the customer name,
trip, total, deposit amount, and bank account already filled in.

## Receiving goods

**Cocokkan Barang** opens the matching screen. It is available while the order
is `dp_paid`, `purchasing`, or `arrived`.

Every line defaults to fully received, because that is the normal case. Change
only the lines with a problem.

| Received quantity | Fulfilment status becomes |
|---|---|
| Equal to ordered | Sudah Dibeli (purchased) |
| Between 1 and ordered | Sebagian (partial) |
| Zero | Tidak Ada (unavailable) |

Saving moves the order to `arrived`.

Items marked unavailable still appear on the invoice, labelled as such, so the
customer understands why the total differs from the original order.

## Cancelling an order

**Batalkan** asks for an optional reason and warns you what will happen:

- Goods already bought for this order are released into stock
- Money already received must be returned by recording a **refund**

Cancelling does not automatically refund. That is a separate, deliberate action,
because the money has to physically leave your bank account too.

---

# Invoices

**Invoice** — billing documents sent to customers.

## Two kinds

| Type | Indonesian | Bills for | Issue when |
|---|---|---|---|
| Deposit | DP | Only the deposit amount | Order is confirmed |
| Final | Pelunasan | Full order value minus payments received | Goods have arrived |

Most businesses only issue final invoices and request the deposit through a
WhatsApp message. Both are supported.

## Issuing an invoice

From the order detail page, **Invoice → Terbitkan**.

Choose the type, optionally set a due date (otherwise the default from Settings
applies), and add a note.

Every amount is **copied into the invoice** at the moment it is issued. Editing
the order afterwards does not alter an invoice the customer has already seen. If
the order really changed, void the old invoice and issue a new one.

Issuing a final invoice moves the order to `invoiced`.

## The PDF

Every invoice renders as a PDF containing:

- Your store name, address, phone, and email from Settings
- Invoice number, issue date, due date
- Billing details and shipping address
- Line items with quantity, unit price, and subtotal
- Totals, amount paid, and balance due
- Payment history
- Bank account details, if a balance is outstanding
- Your closing note

Open it with the **PDF** button. It renders on first request and is cached
afterwards.

## Sending to the customer

**Kirim** opens a dialog with the full message text and two buttons.

```
   ┌──────────────────────────────────────────────┐
   │  Message built from your template            │
   │  ──────────────────────────────────────────  │
   │  Halo Rina, barang pesananmu sudah sampai…   │
   │  Invoice INV-2026-0001                       │
   │  Total: Rp335.000                            │
   │  Sisa pelunasan: Rp167.500                   │
   │  Transfer ke: BCA 1234567890                 │
   └──────────────────────────────────────────────┘
        │                    │
        ▼                    ▼
   [Salin teks]      [Buka WhatsApp]
    copy to           opens wa.me with the
    clipboard         message pre-filled
```

You press send yourself, from your own number. There is no paid gateway and no
unfamiliar sender ID, so customers see a message from a contact they recognise.

Once you open WhatsApp, the invoice is marked as sent and the channel is
recorded.

## Invoice statuses

| Status | Indonesian | Meaning |
|---|---|---|
| `draft` | Draft | Created, not yet sent |
| `sent` | Terkirim | Delivered to the customer |
| `paid` | Lunas | Fully paid |
| `void` | Dibatalkan | Cancelled, replaced by another invoice |

A paid invoice cannot be voided. The money has already been recorded.

## Voiding an invoice

**Batalkan** — the red ban icon on the invoice list, or the **Batalkan** button in
the order's invoice panel. It appears only for invoices that are neither paid nor
already void.

Use it when an invoice was issued with the wrong figures: the customer added an
item after you billed them, or a quantity was corrected. Do not delete and
re-issue silently — the customer may already be holding the old PDF, and the
audit trail should show that it was withdrawn.

Voiding is deliberately narrow in effect:

| What happens | What does not |
|---|---|
| Status becomes `void` | Payments already recorded are untouched |
| The invoice stops counting as an outstanding bill | The order status stays where it is |
| The PDF stays downloadable as a record | Nothing is deleted |

Because the order status does not move, a replacement invoice can be issued
straight away from the same panel.

A confirmation box states all of this before the action runs.

## The invoice list

Search by invoice number, order number, or customer. Filter by status or type.
Invoices past their due date and still unpaid are flagged **lewat tempo**
(overdue).

---

# Shipments

**Pengiriman** — packages and their JNE tracking numbers.

## Step 1 — Pack

From the order, **Tandai Dikemas**.

| Field | Indonesian | Notes |
|---|---|---|
| Courier | Kurir | Defaults to JNE |
| Service | Layanan | REG, YES, OKE, or JTR |
| Weight | Berat | In grams, as weighed on the scale |
| Dimensions | Dimensi paket | Length × width × height in cm; optional |
| Notes | Catatan kemasan | Bubble wrap, fragile, do not stack |

The order moves to `packed`.

### Estimating the shipping cost

**Hitung Ongkir** works out what the courier will charge, using the destination
city already on the order — you do not retype the address.

Couriers bill on whichever is greater: actual weight or **volumetric weight**, a
stand-in for the space the box occupies in the truck.

```
                     length × width × height (cm)
   volumetric kg  =  ────────────────────────────
                              divisor

   chargeable kg  =  max(actual kg, volumetric kg), rounded UP to a whole kg
   cost           =  chargeable kg × rate per kg for the destination city
```

The divisor is 6000 by default, which is what JNE uses. Change it under
Settings → Ongkir if your courier uses 5000.

**Worked example.** A box of instant noodles: 40 × 30 × 25 cm, weighing 800 g.

```
   volumetric  =  (40 × 30 × 25) ÷ 6000  =  30000 ÷ 6000  =  5 kg
   actual      =  0.8 kg
   chargeable  =  max(0.8, 5)            =  5 kg
   cost        =  5 × Rp28.000 (Bandung, YES)  =  Rp140.000
```

Weighing alone would have suggested 1 kg and Rp28.000 — a Rp112.000 shortfall
absorbed out of your margin. This is the single most common way jastip shipping
costs go wrong.

The reverse case: 2.3 kg of skincare in a small box rounds **up** to 3 kg,
because couriers do not sell fractions of a kilogram.

The result panel shows both weights, which one was charged, and the rate used.
If the destination city has no rate on file it says so and falls back to the
default rate — the number is still usable, just less precise.

Pressing **Simpan** stores the dimensions and the estimate on the shipment, so
the figure is there later when you enter the real cost.

## Step 2 — Enter the tracking number

**Input Resi & Kirim**.

| Field | Indonesian | Notes |
|---|---|---|
| Tracking number | Nomor resi | Required, stored uppercase |
| Shipping cost | Ongkir dibayar | What **you** paid JNE, not what you charged |
| Ship date | Tanggal kirim | Defaults to today |

If you estimated the cost at packing time, a **Pakai estimasi** link under the
cost field fills it in. Replace it with the real figure from the receipt when it
differs — the estimate is a starting point, not the record.

Keep the two shipping figures separate. The customer-facing fee is on the order;
the amount you paid the courier is here. The difference is real margin, and the
Profit tab reports both.

### The unpaid guard

An order with an outstanding balance is refused, with the exact amount shown.

To override, tick **Kirim walau belum lunas**. Use it only for regulars you
trust to pay on delivery. The override is recorded in the audit trail.

Shipping freezes the order: its items can no longer be edited.

## Step 3 — Notify the customer

**Kabari Customer** builds a WhatsApp message containing the courier, service,
and tracking number, plus the JNE tracking link. Same flow as invoices: you
press send.

The shipment records when the customer was notified. The shipment list flags
packages that were shipped but never announced, so nobody is left wondering
where their parcel is.

### The cost column on the shipment list

The **Pengiriman** list shows the cost actually paid. Packages not yet handed to
the courier show the estimate instead, labelled *estimasi*, so it is never
mistaken for a final figure when totalling costs.

## Step 4 — Close it out

**Tandai Diterima** sets the shipment to delivered and the order to `completed`.

## Shipment statuses

| Status | Indonesian | Meaning |
|---|---|---|
| `packing` | Dikemas | Being packed |
| `ready` | Siap Kirim | Packed, waiting for a tracking number |
| `shipped` | Dikirim | Handed to the courier |
| `delivered` | Diterima | Confirmed received |
| `returned` | Retur | Came back undelivered |

A shipment cannot be marked shipped without a tracking number. The database
itself enforces this, so a customer is never notified about an empty tracking
number.

---

# Packing Queue

**Siap Kemas** — the warehouse worklist.

It answers one question: what do I need to handle today? Switch between four
stages with the dropdown.

```
   arrived  ──▶  packed  ──▶  invoiced  ──▶  paid  ──▶  (shipped)
      │            │             │             │
   pack it     invoice it    chase the      enter the
                             payment        tracking no.
```

| Stage | Indonesian | What to do next |
|---|---|---|
| `arrived` | Siap dikemas | Goods matched, pack them per customer |
| `packed` | Sudah dikemas | Issue the final invoice |
| `invoiced` | Menunggu pelunasan | Wait for or chase the payment |
| `paid` | Siap dikirim | Enter the JNE tracking number |

Each row shows the recipient, destination, item count, and outstanding balance.
**Proses** jumps to the order.

Filter by trip when a shipment lands and you want to clear that trip's backlog
in one pass.

---

# Customers

**Customer** — the people who order from you.

## Fields

| Field | Indonesian | Required | Notes |
|---|---|:---:|---|
| Name | Nama | Yes | |
| WhatsApp number | Nomor WhatsApp | Yes | Any common format accepted |
| Email | Email | No | Enables the email option on messages |
| Instagram | Instagram | No | |
| Address | Alamat | No | Street, number, RT/RW, district |
| City | Kota | No | |
| Province | Provinsi | No | |
| Postal code | Kode Pos | No | |
| Notes | Catatan | No | Packing preferences, landmarks |

A customer code (CUS-0001) is assigned automatically.

## The phone number is the identity

One number, one customer. The same number in any format is refused on save.

This is deliberate: sales-per-customer reporting is grouped by WhatsApp number.
If one person gets recorded twice under slightly different names, their spending
splits across two rows and nobody shows up as the big buyer they actually are.

## Phone number handling

Numbers are normalised to international format so WhatsApp links always work.

| You type | Stored as |
|---|---|
| `081234567890` | `6281234567890` |
| `0812-3456-7890` | `6281234567890` |
| `+62 812 3456 7890` | `6281234567890` |
| `(0812) 3456-7890` | `6281234567890` |
| `81234567890` | `6281234567890` |
| `+81 90 1234 5678` | `819012345678` |

A leading `+` or `00` means the number already carries a country code, so it is
left alone. Without that marker, a number starting with `8` is treated as
Indonesian.

Fill in the address. An order cannot be created without a shipping address, and
having it on the customer saves retyping it every time.

## Price history across trips

The clock icon on each product row opens its price history: what the product cost
on every trip it has appeared in, in that trip's own currency and at that trip's
locked exchange rate.

| Column | Meaning |
|---|---|
| Trip | Trip code, country, departure date, and the rate that was locked |
| Katalog | Cost price entered when building that trip's catalog |
| Beli riil | Average price actually paid at the till |
| Harga jual | Selling price that was published to customers |
| Dibeli / Terjual | Units bought and units sold on that trip |

Read the two cost columns together. If **Beli riil** is consistently above
**Katalog**, the catalog price is being set too optimistically and the markup is
quietly eating the difference.

The same figures appear as a one-line hint when you add the product to a new trip
catalog, with a **Pakai harga ini** link. The link only appears when the previous
trip used the **same currency** — the same brand bought in Korea instead of Japan
is not a comparable number, and copying it across would be wrong by a factor of a
hundred rather than a few percent.

For that same reason, choosing a product whose master currency differs from the
trip's currency leaves the cost field **empty** rather than pre-filling it.

## Deleting

Deleting hides the customer from lists but keeps the record, because past orders
still point at it. Order history is never lost.

---

# Products

**Produk** — the master catalog.

This is a reference list, not a price list. Actual selling prices live on each
trip's catalog, because exchange rates and shop prices change every journey.

## Fields

| Field | Indonesian | Notes |
|---|---|---|
| Name | Nama produk | Include size or variant |
| SKU | SKU | Auto-generated if left empty |
| Category | Kategori | For filtering and reports |
| Brand | Brand | |
| Usual store | Toko langganan | Helps the tripper find it |
| Currency | Mata uang | Currency of the reference price |
| Reference cost | Harga modal | Typical purchase price abroad |
| Markup type | Jenis markup | Default when adding to a trip catalog |
| Markup value | Markup | Default markup |
| Weight | Berat | Grams, for shipping estimates |
| Image URL | URL gambar | |
| Notes | Catatan | |
| Active | Produk aktif | Inactive products cannot be added to catalogs |

The reference cost and markup are **defaults**, pre-filled when you add the
product to a trip. You can override both per trip.

## Categories

**Kategori** opens a small manager for adding, renaming, and deleting
categories. A category still in use by products cannot be deleted.

## Deleting

Deleting deactivates the product rather than erasing it. It disappears from
catalog pickers, but every past order, purchase, and report stays intact.

---

# Stock

**Stok** — goods you own that nobody ordered.

## Where stock comes from

```
   Tripper buys 8   ──▶   5 go to customer orders
                          3 have no owner
                               │
                               ▼
                          ┌──────────┐
                          │  STOCK   │
                          └──────────┘
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                ▼
        sold on a       released from      corrected by
        marketplace     a reduced order    stock count
```

Three things create stock: surplus purchases, quantity reductions on orders that
were already bought for, and manual adjustments.

## Moving average cost

Each product carries one average cost, recalculated every time stock arrives.

```
   Existing:  3 units @ Rp100,000  =  Rp300,000
   Incoming:  2 units @ Rp110,000  =  Rp220,000
                                      ──────────
   Now:       5 units                  Rp520,000
                                      ──────────
   New average cost = Rp520,000 ÷ 5 = Rp104,000 per unit
```

Average cost is used rather than tracking each batch separately, because jastip
goods are identical units and nobody labels which trip a bottle came from. It
keeps the stock value sensible when purchase prices differ between trips.

## Recording a marketplace sale

**Jual** on a stock row.

| Field | Indonesian | Notes |
|---|---|---|
| Quantity sold | Jumlah terjual | Cannot exceed stock on hand |
| Selling price | Harga jual per pcs | |
| Channel | Kanal penjualan | Shopee, Tokopedia, Instagram |
| Notes | Catatan | |

The margin per unit is shown live as you type, using the moving average cost.

Marketplace sales are recorded separately from trip profit. A trip's report
covers goods that customers ordered; stock sales are a different line of
business with a different margin.

## Stock adjustment

**Sesuaikan** sets the quantity to whatever you physically counted. The
difference is written to the movement history with your reason, so a shrinking
stock level always has an explanation attached.

## Movement history

The **Riwayat Pergerakan** tab logs every change.

| Type | Indonesian | Direction |
|---|---|---|
| `in_purchase` | Masuk dari belanja | + |
| `out_order` | Dipakai pesanan | − |
| `out_marketplace` | Terjual marketplace | − |
| `adjustment` | Penyesuaian | + or − |

---

# Reports

**Laporan** — up to five tabs, each exportable to CSV.

Owners see all five. Admins see everything except Order profit; that margin tab
is hidden from them.

## Receivables

Every order with money outstanding, oldest first.

| Column | Indonesian | Notes |
|---|---|---|
| Order | Order | Click to open |
| Customer | Customer | Includes a "chase via WhatsApp" link |
| Total | Total | Order value |
| Paid | Sudah bayar | Received so far |
| Outstanding | Sisa | Still owed |
| Age | Umur | Days since the order date |

Anything past 14 days is shown in red. Work this list from the top: the oldest
debts are the hardest to collect.

## Order profit

Margin per order. Owner only; the tab does not appear for admins.

| Column | Meaning |
|---|---|
| Revenue | Order total |
| Cost of goods | Real purchase cost allocated to this order |
| Profit | Revenue − cost of goods |
| Margin | Profit ÷ revenue × 100 |

An order showing zero cost of goods has no purchases recorded against it yet.
That is a data gap, not a 100% margin.

Negative profit usually means the shop price rose after you published the
selling price. Compare against the trip catalog to confirm.

## Product performance

Units sold, order count, revenue, cost, and profit per product. Filter by trip.

Use it to decide what to carry next trip and where your markup is too thin.

## Per Customer

One row per customer, biggest spender first. Filter by trip to see who bought on
a particular run, or leave it on all trips for lifetime figures.

| Column | Indonesian | Meaning |
|---|---|---|
| Customer | Customer | Name, code, city, and date of their last order |
| Orders | Order | Number of non-cancelled orders |
| Pieces | Pcs | Total quantity across those orders |
| Revenue | Omzet | Sum of order totals |
| Average | Rata-rata | Revenue ÷ orders |
| Profit | Profit | Revenue − real cost of goods |
| Outstanding | Piutang | Still owed across all their orders |

Two ways to use it. Read from the top to decide who gets priority when a trip
has limited slots or a product has a quota. Then read the Outstanding column: a
customer who buys often *and* owes often is a different problem from one who
simply buys a lot.

A customer with high revenue but low profit is buying your thinnest-margin
products. That is worth knowing before you offer them a discount.

## Per Channel

Where the orders come from, based on the **Asal order** field.

| Column | Indonesian | Meaning |
|---|---|---|
| Channel | Channel | WhatsApp, Instagram, TikTok, Marketplace, or Lainnya |
| Orders | Order | Non-cancelled orders from that channel |
| Customers | Customer | Distinct customers who ordered through it |
| Revenue | Omzet | Sum of order totals |
| Average | Rata-rata | Revenue ÷ orders |
| Profit | Profit | Revenue − real cost of goods |
| Share | Porsi omzet | That channel's percentage of total revenue, drawn as a bar |

The shares always add up to 100%, so the bars are directly comparable.

Read it against the effort each channel costs you. A channel with few orders but
a high average order value may be worth more than a noisy one that produces
small orders — the Average column tells you which is which.

Orders recorded before this field existed default to WhatsApp, so early figures
lean that way until enough new orders accumulate.

## CSV export

Each tab has an **Ekspor CSV** button. Files open cleanly in Excel, including
accented characters in customer names.

---

# Settings

**Pengaturan** — owner only for editing. Five tabs.

## Toko (Store identity)

Appears on invoice PDFs and in customer messages.

| Setting | Indonesian | Used in |
|---|---|---|
| Store name | Nama toko | Invoice header, message templates |
| WhatsApp number | Nomor WA toko | Invoice |
| Email | Email toko | Invoice |
| Address | Alamat toko | Invoice |
| Bank account | Rekening pembayaran | Invoice and every payment request |
| Closing note | Catatan penutup invoice | Invoice footer |
| Invoice due days | Jatuh tempo invoice | Default due date on new invoices |

Get the bank account right. It appears on every invoice and payment request, and
a typo here means money lands nowhere.

## Template Pesan (Message templates)

Three editable templates. Placeholders in `{{double_braces}}` are replaced with
real data when a message is built.

| Template | Indonesian | Used when |
|---|---|---|
| Deposit request | Pesan permintaan DP | Pressing **Tagih DP** |
| Invoice | Pesan penagihan pelunasan | Sending an invoice |
| Shipment | Pesan informasi pengiriman | Notifying about a tracking number |

### Available placeholders

| Placeholder | Replaced with |
|---|---|
| `{{customer_name}}` | Customer name |
| `{{store_name}}` | Your store name |
| `{{trip_title}}` | Trip title |
| `{{order_number}}` | Order number |
| `{{invoice_number}}` | Invoice number |
| `{{total}}` | Formatted total, e.g. Rp335.000 |
| `{{dp_amount}}` | Deposit required |
| `{{amount_paid}}` | Amount already paid |
| `{{amount_due}}` | Outstanding balance |
| `{{due_date}}` | Invoice due date |
| `{{bank_account}}` | Bank account from store settings |
| `{{courier}}` | Courier name |
| `{{service}}` | Service level |
| `{{tracking_number}}` | Tracking number |
| `{{recipient_name}}` | Recipient on the shipping address |

An unrecognised placeholder is left in the message as-is, so a typo shows up
immediately instead of silently disappearing.

## Ongkir (Shipping)

Two parts: how the cost is calculated, and the rate table it is calculated
against.

### Calculation settings

| Setting | Indonesian | Default | Meaning |
|---|---|---|---|
| Volumetric divisor | Pembagi berat volume | 6000 | Divides length × width × height in cm to give volumetric kilograms |
| Fallback rate | Tarif cadangan per kg | Rp25.000 | Used when the destination city is not in the rate table |

JNE, SiCepat, and J&T all use 6000 for domestic parcels. Some couriers and most
international air freight use 5000, which makes every bulky package cost more.
Change it only if your courier's terms say so.

### Rate table

One row per courier, service, and destination city. When you estimate a cost,
the order's shipping city is matched here first; only if there is no match does
the fallback rate apply.

| Field | Indonesian | Notes |
|---|---|---|
| City | Kota tujuan | Matched case-insensitively; "Kota Bandung" and "bandung" are the same row |
| Province | Provinsi | Display only, to tell similarly named cities apart |
| Courier | Kurir | JNE by default |
| Service | Layanan | REG, YES, OKE, or JTR |
| Rate per kg | Tarif per kg | What the courier charges per kilogram |
| Minimum weight | Berat minimum | Usually 1000 g — couriers bill at least 1 kg |
| Estimated delivery | Estimasi tiba | Free text shown with the estimate, e.g. "2-3 hari" |

Saving a city that already exists for the same courier and service **updates**
that row rather than creating a duplicate, so keeping rates current is a matter
of re-entering the city with the new price.

The system ships with rates for the cities Indonesian jastip shops send to most
often. Add yours as you encounter them: an estimate against the fallback rate is
a guess, one against a real rate is a number you can quote to a customer.

Editing rates is restricted to owners; admins can see them and use the estimate.

## Pengguna (Users)

Add, edit, deactivate, and delete accounts, and reset passwords.

| Action | Notes |
|---|---|
| Add user | Name, email, password, role |
| Edit user | Email cannot be changed; use deactivate and create instead |
| Reset password | Signs the user out of every device |
| Deactivate | Blocks login without deleting history |
| Delete | Permanent; audit trail entries survive |

Two safeguards apply: you cannot delete your own account, and you cannot remove
or demote the last active owner. Otherwise nobody could manage users any more.

## Jejak Perubahan (Audit trail)

Who changed what, and when. Filter by entity type.

| Action | Indonesian | Recorded for |
|---|---|---|
| `create` | Dibuat | New orders, purchases, invoices |
| `update` | Diubah | Order edits, settings changes |
| `item_change` | Ubah item | Quantity and price changes on orders |
| `status_change` | Ubah status | Order and trip status moves |
| `payment_record` | Catat pembayaran | Payments recorded |
| `delete` | Dihapus | Deleted payments and purchases |
| `ship` | Kirim | Tracking numbers entered |

This is the screen to open when a total does not look right and you need to know
who changed the quantity, and when.

---

# Appendix A — Money and rounding

All amounts are stored with two decimal places and calculated using exact
decimal arithmetic, never floating point. A one-rupiah drift across a hundred
orders would show up in the profit report, so it is designed out.

| Situation | Rule |
|---|---|
| Currency conversion | Multiply by the rate, round to whole rupiah |
| Published selling price | Round **up** to the nearest Rp100 |
| Deposit percentage | Round to whole rupiah |
| Report totals | Round to whole rupiah for display |

Displayed as `Rp1.250.000`, using Indonesian thousands separators.

---

# Appendix B — Worked example, end to end

A complete trip with real numbers, so you can check the system against your own
arithmetic.

**Setup.** Trip to Japan, rate ¥1 = Rp100.

| Product | Cost | Markup | Selling price |
|---|---|---|---|
| Lotion | ¥1,000 | 30% | Rp130,000 |
| Snack Box | ¥500 | +Rp25,000 | Rp75,000 |

**Orders.**

| Order | Customer | Items | Total |
|---|---|---|---|
| ORD-0001 | Rina | 2 × Lotion, 1 × Snack Box | Rp335,000 |
| ORD-0002 | Budi | 3 × Lotion | Rp390,000 |

**Deposits.** Rina pays Rp167,500 (50%). Budi pays Rp195,000 (50%).

**Purchases.** The tripper buys 8 Lotion at ¥1,000 and 1 Snack Box at ¥500.

```
   8 Lotion  ──▶  5 to orders (2 Rina + 3 Budi),  3 to stock
   1 Snack   ──▶  1 to order (Rina),              0 to stock
```

**A change.** Budi reduces his order from 3 to 2.

| Effect | Result |
|---|---|
| Order total | Rp390,000 → Rp260,000 |
| Budi's balance | Rp65,000 |
| Stock | 3 → 4 units of Lotion |

**Completion.** Rina's order is received, packed, invoiced, paid in full, and
shipped. Budi's is still outstanding.

**Trip expenses.** Extra baggage Rp850,000, airport train Rp150,000.

**The report.**

| Line | Calculation | Amount |
|---|---|---|
| Revenue | 335,000 + 260,000 | **Rp595,000** |
| Cost of goods | 4 Lotion × 100,000 + 1 Snack × 50,000 | **Rp450,000** |
| Gross profit | 595,000 − 450,000 | **Rp145,000** |
| Trip expenses | 850,000 + 150,000 | **Rp1,000,000** |
| **Net profit** | 145,000 − 1,000,000 | **−Rp855,000** |

| Cash line | Calculation | Amount |
|---|---|---|
| Total capital out | (8 × 100,000) + (1 × 50,000) + 1,000,000 | Rp1,850,000 |
| Surplus stock | 4 units × Rp100,000 | Rp400,000 |
| Outstanding | Budi's remaining balance | Rp65,000 |

Note that cost of goods counts 4 units of Lotion (2 for Rina, 2 for Budi), not
the 8 that were bought. The other 4 sit in stock worth Rp400,000 and are not
charged against this trip.

This trip lost money because Rp1,000,000 of expenses was spread over only two
small orders. That is the correct conclusion: on a real trip those fixed costs
are spread over far more orders.

---

# Appendix C — Common tasks

| I want to… | Go to |
|---|---|
| Start a new trip | Trip → Buat Trip |
| Set prices for a trip | Trip → open it → Katalog |
| Record an order from WhatsApp | Order → Catat Order |
| Ask a customer for their deposit | Order detail → Tagih DP |
| See what to buy abroad | Daftar Belanja |
| Record something I just bought | Daftar Belanja → Catat Beli |
| Check goods against orders | Order detail → Cocokkan Barang |
| Find today's packing work | Siap Kemas |
| Bill a customer | Order detail → Invoice → Terbitkan |
| Enter a JNE tracking number | Order detail → Input Resi & Kirim |
| Find out who owes me money | Laporan → Piutang |
| See if a trip made money | Trip → open it → Profit |
| Sell leftover goods | Stok → Jual |
| Change the invoice bank account | Pengaturan → Toko |
| Give the tripper an account | Pengaturan → Pengguna |
| Find out who changed an order | Pengaturan → Jejak Perubahan |

---

# Appendix D — Glossary

| Indonesian | English | Meaning here |
|---|---|---|
| Jastip / jasa titip | Personal shopping service | Buying abroad on request |
| Tripper | Traveller | The person who travels and buys |
| DP (uang muka) | Deposit | Partial payment that confirms an order |
| Pelunasan | Settlement | The final payment |
| Resi | Tracking number | Courier tracking reference |
| Ongkir | Shipping fee | Delivery charge |
| HPP | COGS | Cost of goods sold |
| Omzet | Revenue | Gross sales |
| Laba kotor | Gross profit | Revenue minus cost of goods |
| Laba bersih | Net profit | Gross profit minus expenses |
| Piutang | Receivables | Money customers owe you |
| Kurs | Exchange rate | Foreign currency to rupiah |
| Markup | Markup | Amount added to cost to set the price |
| Stok | Stock | Goods on hand with no owner |
| Kuota | Quota | Maximum units offered on a trip |
