# The mythology behind the names

Grazhda, zgard, dukh and molfar are not invented words. They come from Hutsul
culture — the highlanders of the Ukrainian Carpathians — and each one carries a
meaning that turns out to describe a piece of software rather well.

This document explains where the names come from, why each fits the component it
labels, and offers a lexicon for naming the components that do not exist yet. It
also argues for some discipline about when *not* to reach for a Hutsul word.

Where I am reporting folklore or ethnography I say so. Where I am proposing a
mapping onto software, that is my suggestion, not tradition.

---

## Ґражда — grazhda — the enclosed homestead

**Pronunciation:** *GRAZH-da* (the ґ is a hard g, as in "go").

A grazhda is the traditional Hutsul farmstead of the Carpathians: a dwelling and
its outbuildings — byre, granary, stable, workshop — joined under continuous
roofs so that they form a closed square around an inner courtyard. The outer
walls are blank; there is one gate. Everything the household needs is inside, and
the mountain is outside.

They were built that way for reasons that were entirely practical. The
Carpathians had bears and wolves, winters that buried a house to the eaves, and
for long stretches of history the *opryshky*, the highland outlaws. A grazhda
meant a family could shut one gate and be self-sufficient until spring.

The word shares a root with Slavic *gorod* / *hrad* — an enclosure, a fortified
place — and, further back, with English *yard* and *garden*. The idea underneath
all of them is the same: **a boundary drawn around everything that belongs
together.**

**Why it names the project.** A grazhda is what the tool builds for you: an
enclosure containing every repository, every tool version, every configuration
you need to work, with one gate and nothing scattered across the mountain. The
existing design already reflects this without anyone having to argue for it —
packages install into the project's own directory rather than contaminating the
host system, workspaces are declared rather than discovered, and the whole
working world is described in one file. That is a grazhda: bounded,
self-sufficient, yours.

---

## Згарда — zgard — the linked amulet

**Pronunciation:** *ZGAR-da*.

A zgarda is a Hutsul neck ornament: rows of small cast brass or bronze crosses
threaded onto a cord or wire, often separated by spiral tubes of coiled brass.
Women wore them, sometimes several rows deep, and they were both jewellery and
amulet — protective, and often inherited, with individual crosses acquired over a
lifetime and added to the string.

Two things about it matter for the metaphor. First, **it is many small pieces on
one thread** — the value is not in any single cross but in the strand that holds
them in relation to one another. Second, **it is apotropaic**: it is worn to keep
something bad from happening.

**Why it names the CLI.** `zgard` is the command that operates on many
repositories as one thing. Thirty repositories are the crosses; the workspace
configuration is the cord. And its most-used features are protective ones —
`--dry-run`, the confirmation prompt, the refusal to purge without an explicit
target, the warning when a filter matches more than you probably meant. A tool
that fans a command out across thirty repositories needs to be an amulet as much
as an instrument.

---

## Дух — dukh — the spirit of the house

**Pronunciation:** *dooKH*, with the final sound like the *ch* in Scottish
"loch".

*Dukh* simply means spirit, or breath — the same root as *dykhaty*, to breathe.
In Slavic folk belief the household had its own resident spirit, the *domovyk*,
who lived behind the stove, watched over the livestock, and made his displeasure
known when the household grew disorderly: noises at night, tools moved,
animals unsettled. He was not a threat. He was the house paying attention to
itself.

There is a happy coincidence here. The Unix word *daemon* comes from the Greek
δαίμων, which in Plato and later usage meant a benevolent attendant spirit — a
guiding presence rather than a devil. The MIT programmers who named the first
background processes meant it in exactly that sense.

**Why it names the daemon.** `dukh` is the process that lives inside the grazhda
and pays attention when nobody is looking. It notices when a repository has
drifted from where it should be and says so. It is the house noticing its own
disorder. Both the Slavic *domovyk* and the Greek *daemon* land on the same idea,
which is a rare thing in naming.

---

## Мольфар — molfar — the one who reads the weather

**Pronunciation:** *MOL-far*.

A molfar is the Hutsul folk magician: healer, weather-worker, seer. The
ethnographic record credits them with turning away hailstorms, finding lost
people and livestock, treating illness, and foretelling. The knowledge was held
to pass by apprenticeship — a molfar took a student and taught them over years —
and there was a moral distinction between those who used it for the community and
those who used it against people.

They are widely known in Ukrainian culture through Mykhailo Kotsiubynsky's
*Shadows of Forgotten Ancestors* and Sergei Parajanov's 1965 film of it.

What distinguishes a molfar from the other figures here is **time horizon and
judgement**. The zgarda protects in the moment. The dukh notices continuously.
The molfar has accumulated knowledge over years, reads signs, decides what they
mean, and acts before the storm arrives.

**Why it names the long-term process.** `molfar` is the component with memory and
judgement: it remembers what was tried, plans work, runs it across the enclosure,
records what happened, and brings back a proposal. And the folklore contains its
own governance principle — the distinction between a molfar who serves the
community and one who does not is exactly why the design insists that the system
*proposes* and a human *approves*. An automated process acting on your
repositories without oversight is the wrong kind of molfar.

The name was already latent in the old project: `guides/feature-ideas.md` files
its most ambitious future work under "Phase 3+ (Molfar vision)".

---

## A lexicon for what comes next

These are real Hutsul and Carpathian terms with — in my view — obvious software
counterparts. Mappings are proposals, not tradition.

### Strong candidates

| Word | Ukrainian | What it is | Suggested for |
| :--- | :--- | :--- | :--- |
| **trembita** | трембіта | The alpine horn, three to four metres long, sounded to carry news across valleys: a death, a danger, the herds moving | The notification and event system. Nothing else in the lexicon is so exactly a message bus |
| **vatra** | ватра | The fire kindled when the herds go up to the summer pasture and kept alive until they come down; letting it die was a disgrace | The always-on service, or the append-only run journal — the thing that must not go out |
| **strunka** | струнка | The narrow gate on the pasture through which sheep pass one at a time to be milked | The work queue: a serialising pipeline where each item passes through in turn |
| **vatah** | ватаг | The head shepherd who runs the summer pasture — allocates duties, keeps the accounts, leads the season | The scheduler and orchestrator inside `molfar` |
| **koliba** | коліба | The shepherds' temporary hut on the high pasture, built for a season and left | An ephemeral sandbox or throwaway git worktree where a plan is executed before anyone approves it |
| **cheres** | черес | The wide Hutsul leather belt, pocketed, holding tools and money | The package registry — the toolbelt the enclosure carries |

### Worth considering

| Word | Ukrainian | What it is | Suggested for |
| :--- | :--- | :--- | :--- |
| **polonyna** | полонина | The high summer pasture the herds are driven to | A remote or cloud workspace, as opposed to the local grazhda |
| **bartka** | бартка | The Hutsul axe-staff: walking stick, tool, weapon and mark of status | A deterministic code transformation — a codemod |
| **opryshky** | опришки | The Carpathian outlaws who raided the valleys; Oleksa Dovbush is the famous one | The adversarial test suite or fault-injection harness — the thing that attacks your enclosure so you find out whether the walls hold |
| **chuhaister** | чугайстер | A benevolent forest giant who protects travellers and hunts the forest spirits that lure them | The guard that treats repository content as untrusted input (see `IDEAS.md` §11) |
| **didukh** | дідух | The sheaf of wheat brought into the house at midwinter, standing for the ancestors — note that it contains *dukh* | Archival and retention: what is kept when the records age out |

### One that is too good to leave out

| Word | Ukrainian | What it is | Suggested for |
| :--- | :--- | :--- | :--- |
| **mavka** | мавка / нявка | A forest spirit in the shape of a beautiful young woman who lures travellers off the path — and whose back, seen from behind, is hollow, with the organs exposed | A plausible-looking but hollow result |

That last one is not a joke. The characteristic failure of a language model is
work that is convincing from the front and empty behind: a patch that compiles
and reads well and does not do what was asked. Hutsul folklore already has a
precise word for that, and it happens to describe the single thing your
verification layer exists to catch.

If you take the suggestion in `IDEAS.md` §5 that every plan step must carry an
oracle, then the component that evaluates those oracles has an obvious name, and
"this step failed the mavka check" is an unusually clear way to say that
something looked right and was not.

---

## Naming discipline

The temptation with a lexicon this good is to use all of it. Resist. Some rules
that will keep the project usable by people who do not speak Ukrainian.

**Proper nouns, not verbs.** Mythological names go on things you name once and
refer to often: processes, binaries, long-lived subsystems. Commands and flags
stay plain English — `zgard ws init`, `--dry-run`, `--parallel`. A newcomer can
learn four proper nouns in a minute. They cannot learn forty verbs in a language
they do not read, and they should not have to.

**One gloss, everywhere.** Every mythological name gets a one-line explanation
wherever it first appears: in `--help`, in the README, in the log line where the
process starts. `dukh — the house spirit; watches your workspaces` costs six
words and removes the entire objection.

**Searchability matters.** A contributor hitting an error will paste it into a
search engine. Make sure the error text contains an English word describing what
failed as well as the component name: `dukh (workspace monitor): cannot bind
:50501` rather than `dukh: bind failed`.

**Do not rename what people already know.** The daemon is a daemon, the CLI is a
CLI, a pull request is a pull request. The names label *your* components, not
the industry's concepts.

**Transliteration, once.** Pick one spelling per name and never vary it. `zgard`
not `zgarda`/`zharda`; `dukh` not `duh`/`dux`; `molfar` not `mol'far`. Put the
Ukrainian spelling in this document and use the Latin form everywhere else.

---

## Why any of this is worth the effort

Two reasons, one soft and one hard.

The soft one: a project with a coherent naming world is a project people
remember and enjoy working on. "The house spirit watches the enclosure and the
molfar plans the work" is a sentence that teaches the architecture in passing.

The hard one: these particular metaphors turned out to be load-bearing. The
grazhda's enclosure is why packages install into the project's own directory
instead of the host system. The zgarda's cord is why the workspace configuration
is one file rather than thirty. The molfar's moral distinction is the argument
for the approval gate. When a metaphor is chosen well it stops being decoration
and starts making architectural decisions for you — and when a proposed feature
does not fit the metaphor, that is worth a second look before you build it.

---

## Sources and honesty about them

The ethnographic detail here — the construction of the grazhda, the composition
of the zgarda, the role of the molfar and the vatah, the trembita's signalling
use — is drawn from general knowledge of Hutsul material culture and Carpathian
folklore, of the kind found in Ukrainian ethnographic literature and in
Kotsiubynsky's *Shadows of Forgotten Ancestors* and its 1965 film adaptation.

I have not verified individual claims against primary sources, and folk practice
varied considerably between valleys and across the centuries; treat the
specifics as broadly right rather than precisely documented. If you want to use
any of this publicly — in a README or a talk — it is worth an hour with a proper
ethnographic source first, and the Hutsul museums in Kolomyia and Verkhovyna are
the obvious starting points.

Every mapping from a Hutsul term to a software component is mine, and is offered
as a suggestion.
