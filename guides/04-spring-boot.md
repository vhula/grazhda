# Spring Boot 4.1, for a Java developer building molfar

You write Java every day and you know Maven. You have never built a Spring Boot application. You are about to build `molfar`, the long-running process at the centre of **grazhda** — a tool that automates the work of a software engineer.

This guide teaches Spring from zero and assumes everything you already know about Java and about builds. It will not explain what an interface is, what an annotation is, or what a `<dependency>` element does. It will explain, at length, why a framework that claims to remove boilerplate can fail in complete silence, why the most important structural decision in this service is to keep Spring out of most of it, and why an annotation you can see on a method sometimes does nothing at all.

Everything version-specific here targets **Java 25 and Spring Boot 4.1**, released 10 June 2026. That matters more than usual. Spring Boot 4 is a genuine structural break from Spring Boot 3, and the internet — including a great deal of otherwise good writing — is still full of Boot 3 instructions that are now incomplete rather than obviously wrong. Where a claim is version-sensitive and I could not verify it against current documentation, this guide says so instead of guessing. When this guide and the artifact Maven actually resolves disagree, believe Maven.

What molfar has to do, so the examples have somewhere to land: accept a **chore** submitted by an engineer — an intent, a scope of repositories, a budget — plan it into ordered steps, execute them across git repositories on the local machine, and return a **proposal** the human approves, never pushing anything unapproved; persist chores, plans, runs, per-step outcomes, proposals and approvals in a **durable journal** that survives restarts; run **recurring scheduled jobs** (a nightly dependency sweep, a morning health report); consume a **gRPC contract** from `dukh`, the Go daemon that reports repository health; expose a **REST API** described by a contract-first OpenAPI document and consumed by the `zgard` CLI; and run for weeks on an engineer's laptop with no external infrastructure.

---

# Part 1 — What Spring Boot is, and why this process needs it

## 1.1 What a long-running process needs

`zgard` is a CLI: it starts, does one thing, prints, exits. `molfar` starts once and stays up for weeks. That difference generates almost every requirement in this guide.

A process that stays up needs a **lifecycle** — something to construct the object graph at start-up in dependency order, hold it, and tear it down in reverse when a `SIGTERM` arrives. It needs **scheduling**, because the nightly sweep must fire at 02:00 whether or not anyone is watching, on a thread pool that outlives every request. It needs **persistence with transactions**, because a chore's plan, its runs and its per-step outcomes must land atomically and be there after a restart. It needs an **HTTP server**, because `zgard` talks to it. It needs **configuration** that differs between a laptop and a test run without recompiling. And it needs **observability** — health, metrics, structured logs — because a process you cannot see inside is one you will eventually restart out of superstition. None of that is hard in isolation. All of it is tedious, and all of it is the same tedium in every service ever written.

## 1.2 What you would hand-roll

You would write a `main` constructing forty objects in the right order and holding them so they are not collected. You would add a shutdown hook, then discover that ordering matters: closing the datasource before stopping the scheduler means the last nightly sweep dies with "pool is closed" instead of recording an outcome. You would embed Tomcat and write your own routing from method-plus-path to handler, plus JSON binding both ways, plus a way to turn an exception into a status code that is not `500`. You would build a connection pool and then a transaction manager binding a connection to the current thread — because every repository method that participates in a caller's transaction must find the same connection without being passed it. You would write scheduling and a cron parser, and a properties loader layering a file over environment variables over command-line arguments, with defined precedence, type conversion and validation.

Each is a day; together they are a month with no relationship to what grazhda is for. Worse, it is *load-bearing* code: that thread-bound transaction binding in particular is the kind of thing that works for six months and then produces one wrong row under concurrency.

## 1.3 What Spring actually is

Strip the marketing and Spring Boot is five things.

**An inversion-of-control container.** You declare what your objects need; the container constructs them and passes the dependencies in. `new` moves out of business code. The container holds each constructed object — a **bean** — for the life of the application, resolves the graph, detects cycles, and closes things on shutdown in reverse construction order. Everything else is built on this.

**Auto-configuration.** Boot inspects the classpath and your configuration and contributes beans conditionally: a JDBC driver plus a `spring.datasource.url` and no `DataSource` you defined yourself means Boot defines one. This is what people mean by "magic", and section 2.2 is about how that magic changed in Boot 4. **Starters** are the curated dependency aggregates that feed it: `spring-boot-starter-webmvc` pulls in Spring MVC, embedded Tomcat, Jackson, and — the part that changed in Boot 4 — the auto-configuration module that wires them together, with a BOM aligning versions so you never write a `<version>` for anything Spring manages.

**An embedded server**, so there is no application server, no WAR and no deployment step: `java -jar molfar.jar` starts a process with Tomcat inside it, which for a laptop service is exactly right. And **a build plugin**, `spring-boot-maven-plugin`, which repackages your jar into an executable one carrying its dependencies and adds a `run` goal for development.

## 1.4 The cost, honestly

A guide that sells Spring as free will be distrusted the first time a context fails to start.

**Memory.** A service with web, JPA and Actuator on it does not fit in 64 MB. Expect a resident set of roughly 250–400 MB untuned, of which maybe 120–200 MB is heap and the rest is metaspace, code cache, thread stacks and the JVM itself. On a laptop already running an IDE and a browser, that is noticeable; section 12.4 covers the floor. **Start-up time** is one to three seconds of component scanning, auto-configuration evaluation, Hibernate metamodel building and pool warm-up — irrelevant for a process that runs for weeks, very relevant for tests, which is why Part 11 is largely about not starting the context.

**The debugging tax.** This is the real cost. When you write wiring by hand, "why did this object get that collaborator" is answered in a file you wrote. Under Spring the answer is distributed across a conditional auto-configuration you have never read, a property you did not set, the order two `@Configuration` classes were processed in, and whether some jar happened to be present. You will spend afternoons on this. Learn the mitigations — the condition evaluation report, `mvn dependency:tree`, the Actuator `conditions` endpoint — early rather than when you are already frustrated. There is a **vocabulary tax** on top: bean, context, proxy, stereotype, slice, starter, advice, `@Primary`, `@Qualifier`, `@ConditionalOnMissingBean`. None of it is hard; all of it is new, and much of it is invisible in your source.

The trade is worth it for molfar, but for a precise reason: not because Spring saves typing, but because persistence, transactions, scheduling and lifecycle are exactly the machinery that is expensive to get subtly wrong — and because the *interesting* logic, planning a chore and executing steps, is going to live outside all of it anyway. That is Part 10, and it is the most important section here.

---

# Part 2 — Spring Boot 4.1 specifically

## 2.1 The version matrix

Boot 4.1 was released 10 June 2026.

| Component | Version |
|---|---|
| Spring Framework | 7.0.8 |
| Spring Security | 7.1.0 |
| Spring Data | 2026.0.0 |
| Hibernate ORM / Validator | 7.1 / 9.1 |
| Tomcat | 11 |
| Jakarta Persistence / Servlet / Validation | 3.2 / 6.1 / 3.1 |
| Jackson | 3.0 |
| Spring gRPC | 1.1.0 |
| OpenTelemetry | 1.62.0 |

The **Java baseline is JDK 17**, unchanged from Boot 3, so Java 25 is comfortably inside range. One dependency has a higher floor: jOOQ 3.20+ needs Java 21. Records, sealed interfaces, pattern matching and virtual threads are all available, and this guide uses records heavily — they are the right shape for configuration properties, request bodies and domain values. For the exact version of anything not in that table, generate a project on `start.spring.io` and read `mvn dependency:tree`; that is authoritative for your exact patch version and takes thirty seconds.

## 2.2 The headline change: the starter is the contract

This is the section that will save you an afternoon. Read it twice.

In Spring Boot 3, all auto-configuration lived in one jar, `spring-boot-autoconfigure`, which had grown to about 2 MiB and contained auto-configuration for every technology Boot had ever supported. Each was guarded by conditions — "only if class `org.flywaydb.core.Flyway` is present". The practical consequence was a simple, learnable rule: **put a library on the classpath and Spring wires it up.**

Spring Boot 4 broke that jar apart. Auto-configuration now ships as focused per-technology modules named `spring-boot-<technology>`: `spring-boot-flyway`, `spring-boot-data-jdbc`, `spring-boot-webclient`. Root packages moved with them, from `org.springframework.boot.autoconfigure.*` to `org.springframework.boot.<technology>.autoconfigure.*`. Starters pull in their modules transitively, so `spring-boot-starter-webmvc` still gets you everything you expect. The problem is what happens when you skip the starter. Suppose you want Flyway migrations and, reasoning from Boot 3 habits, add the library:

```xml
<dependency>
  <groupId>org.flywaydb</groupId>
  <artifactId>flyway-core</artifactId>
</dependency>
```

You write `src/main/resources/db/migration/V1__create_chore.sql` and start molfar. The application starts cleanly. There is **no error**, no warning, no stack trace. Flyway's characteristic banner and `Migrating schema` lines are **absent from the log**. The migration does not run. The first request touching the `chore` table fails with a missing-table error, at which point you go looking at your SQL, your `spring.flyway.locations`, your file naming and your datasource — all of which are fine.

The cause is that `flyway-core` is on the classpath but `spring-boot-flyway`, the module containing `FlywayAutoConfiguration`, is not. No code anywhere has the job of noticing this. As Zakaria Bouazza put it after losing time to exactly this, *"a missing auto-config module fails identically to a missing library — total silence."* The fix is one dependency:

```xml
<dependency>
  <groupId>org.springframework.boot</groupId>
  <artifactId>spring-boot-starter-flyway</artifactId>
</dependency>
```

which resolves as `spring-boot-starter-flyway` → `spring-boot-flyway` (auto-configuration) → `flyway-core` (library). Both halves, one declaration. Generalise it, because Flyway is only the example that caught people first:

> **In Spring Boot 4 the starter is the contract. A raw library gives you the library and nothing else — no wiring, no error, no log line. Silence.**

Anywhere you would have reasoned "the jar is on the classpath, therefore Spring will configure it", add the starter instead. `spring-boot-starter-h2-console`, `spring-boot-starter-flyway`, `spring-boot-starter-liquibase`, `spring-boot-starter-restclient` and `spring-boot-starter-webclient` all exist for exactly this reason. There is a bridge for teams migrating in a hurry: `spring-boot-starter-classic` and `spring-boot-autoconfigure-classic` restore a Boot 3-shaped classpath where everything auto-configures from detection again. Use it to get an upgrade compiling, not as a destination — molfar is new code and should never need it.

## 2.3 Diagnosing silence

Because the failure mode is nothing happening, learn these three now.

**The condition evaluation report.** Set `debug=true` in `application.yaml`, or pass `--debug`, and Boot prints every auto-configuration class it considered under `Positive matches` and `Negative matches`, with reasons. If `FlywayAutoConfiguration` appears in neither list, the class is not on the classpath at all — your missing module, stated plainly.

**The Actuator conditions endpoint.** `GET /actuator/conditions` gives the same report as JSON on a running process, which beats restarting with a flag.

**`mvn dependency:tree`.** Ground truth about what actually resolved. `mvn -q dependency:tree -Dincludes=org.springframework.boot:spring-boot-flyway` returning nothing is the answer.

## 2.4 Renamed starters

Modularisation came with renames. The one you hit immediately:

| Boot 3 | Boot 4 |
|---|---|
| `spring-boot-starter-web` | `spring-boot-starter-webmvc` |
| `spring-boot-starter-web-services` | `spring-boot-starter-webservices` |
| `spring-boot-starter-oauth2-client` | `spring-boot-starter-security-oauth2-client` |
| `spring-boot-starter-oauth2-resource-server` | `spring-boot-starter-security-oauth2-resource-server` |

`spring-boot-starter-web` still publishes an artifact at 4.x, but the reference documentation marks it deprecated in favour of `spring-boot-starter-webmvc`. Write `webmvc` in new code; every tutorial you find will say `web`, it will mostly still work, and you should still not copy it.

Boot 4 also gives most starters a test companion — `spring-boot-starter-webmvc-test`, `spring-boot-starter-data-jpa-test` — carrying that technology's test-slice support, alongside the general `spring-boot-starter-test` which still brings JUnit Jupiter, AssertJ, Mockito and the test context machinery. If `@WebMvcTest` or `@AutoConfigureMockMvc` mysteriously does nothing, a missing test starter is the first suspect: section 2.2's failure shape, one layer over. Confirm the exact set your version publishes against start.spring.io rather than trusting this table forever.

## 2.5 Jackson 3

**Jackson 3.0 is the default in Boot 4, and Jackson 2 support ships deprecated.** Every REST body goes through it, so this matters.

Packages moved: core processing is now under `tools.jackson.*` — `tools.jackson.databind.json.JsonMapper`, for instance — while the *annotations* deliberately stayed at `com.fasterxml.jackson.annotation`, so `@JsonProperty`, `@JsonIgnore` and `@JsonFormat` are unchanged. Any code copied from a Boot 3 tutorial importing `com.fasterxml.jackson.databind.ObjectMapper` will not compile. `ObjectMapper` gives way to `JsonMapper`, built through an immutable builder and locked at `build()`, which removes a class of "someone reconfigured the shared mapper" bugs. `JacksonException` became a `RuntimeException`, so Jackson is usable inside lambdas without ceremony. `JsonSerializer` and `JsonDeserializer` became `ValueSerializer` and `ValueDeserializer`. And default date handling flipped: Jackson 2 wrote numeric timestamps, Jackson 3 writes ISO-8601 strings — a wire-format change you must be deliberate about if `zgard` parses timestamps. There is a `spring.jackson.use-jackson2-defaults` escape hatch for migrations; new code should not use it.

New in 4.1: readers and writers configurable through `spring.jackson.read.*` and `spring.jackson.write.*`, factory limits through `spring.jackson.factory.*`, and `JsonFactoryBuilderCustomizer`, `CborFactoryBuilderCustomizer` and `XmlFactoryBuilderCustomizer` beans for the rest. Practical advice for molfar: records for bodies, `java.time` types, an explicitly set date format rather than any default, and a `@JsonTest` (Part 11) pinning the contract so a Jackson upgrade cannot silently change what `zgard` receives.

## 2.6 What else arrived, and what it means here

**gRPC auto-configuration** is the flagship 4.1 feature and lands directly on molfar's need to consume `dukh`: both server and client sides auto-configured, a standalone Netty transport or a Servlet HTTP/2 transport, `@GrpcAdvice` for centralised exception handling, and automatic observation and metrics. Part 9 covers it properly.

**HTTP Service Clients.** Annotate a plain Java interface with `@HttpExchange` and its methods with `@GetExchange` or `@PostExchange`, register it with `@ImportHttpServices`, and Spring generates the implementation. If molfar ever calls an HTTP API — a package registry, a git host — this replaces both hand-written `RestTemplate` code and third-party declarative clients. Alongside it, **SSRF mitigation** arrives as a configurable `InetAddressFilter` on blocking and reactive HTTP clients: molfar takes a repository scope from an engineer and could be induced to fetch a URL derived from it, so a filter refusing link-local and loopback addresses is a cheap, real defence.

**API versioning auto-configuration** for MVC and WebFlux, driven by `spring.mvc.apiversion.*`, with `ApiVersionResolver`, `ApiVersionParser` and `ApiVersionDeprecationHandler` beans for anything unusual (section 6.5). **`RestTestClient`** for integration tests, against a running server or over `MockMvc` — with one wrinkle: HTTP test clients are no longer auto-configured in Boot 4, so you enable it with `@AutoConfigureRestTestClient` explicitly, and forgetting that means the client will not inject, silence in a different costume.

**An OpenTelemetry starter**, `org.springframework.boot:spring-boot-starter-opentelemetry`, gives vendor-neutral tracing, metrics and log export without requiring Actuator; `management.opentelemetry.enabled` disables the SDK while leaving propagators in place and `management.opentelemetry.tracing.sampler` controls sampling. **Log rotation strategies** for Log4j2 — `size` (default), `time`, `size-and-time`, `cron` — arrive via `logging.log4j2.rollingpolicy.strategy`, which for a process running for weeks on a laptop is not a footnote (section 12.2). And **`@Async` now propagates context across the thread boundary automatically**: in Boot 3 this needed explicit `ContextSnapshot` work, or you lost the trace ID and MDC entries exactly where you most needed them.

Also worth knowing: `spring.config.import` accepts an encoding (`classpath:import.properties[encoding=utf-8]`); the Actuator `info` endpoint exposes process data (`process.uptime`, `process.startTime`, `process.workingDirectory`); and `spring.datasource.connection-fetch: lazy` wraps the datasource in a `LazyConnectionDataSourceProxy`, avoiding a connection for transactions that never touch the database.

## 2.7 Removed and deprecated

**Removed in 4.1:** the deprecated `layertools` jar mode.

**Derby.** The Apache Derby project has been retired; Boot 4.1 deprecates Derby support and `DatabaseDriver.DERBY` and points at H2 or HSQL. Relevant to Part 7's database choice: Derby is not a candidate.

**Devtools LiveReload** is deprecated, as are the Dynatrace V1 API properties.

**A Maven behaviour change that will bite the build.** `-DskipTests` no longer skips AOT processing; use `-Dmaven.test.skip` to skip test compilation and execution entirely. Section 3.4, because grazhda's `just build-service` passes exactly that flag.

---

# Part 3 — The first application

## 3.1 start.spring.io

`https://start.spring.io` generates the project. Choose Maven, Java 25, Spring Boot 4.1.x, group `com.grazhda`, artifact `molfar`, jar packaging. Add **Spring Web** (which resolves to `spring-boot-starter-webmvc`), **Validation**, **Spring Data JPA**, **Flyway Migration**, **Actuator**, **gRPC Client** and **H2 Database**.

Use it rather than writing the `pom.xml` yourself, even though you know Maven perfectly well, for two reasons. It is the authoritative source for which starter artifact ids exist at your exact Boot version, and section 2.4's renames make that a live question. And selecting the gRPC starter makes it add the protobuf Maven plugin for you, which is fiddly by hand. Generate, unzip into `service/`, then read what you got — that is the actual lesson.

## 3.2 The pom, element by element

```xml
<parent>
  <groupId>org.springframework.boot</groupId>
  <artifactId>spring-boot-starter-parent</artifactId>
  <version>4.1.0</version>
  <relativePath/>
</parent>

<groupId>com.grazhda</groupId>
<artifactId>molfar</artifactId>
<version>0.1.0</version>

<properties>
  <java.version>25</java.version>
</properties>
```

The **parent POM** does three jobs. It imports `spring-boot-dependencies`, a BOM with managed versions for several hundred artifacts, which is why no dependency below carries a `<version>`. It configures plugin defaults — compiler release, resource filtering, surefire. And it defines the `java.version` property, which it translates into `maven.compiler.release`; setting `<java.version>25</java.version>` is therefore the whole of "target Java 25", and you do not also configure `maven-compiler-plugin`.

The empty `<relativePath/>` tells Maven not to look for the parent on disk. Without it Maven checks `../pom.xml`, and in grazhda's layout — `service/` under a repository root with no parent POM — that check is pure noise. If your organisation forbids an external parent, import the BOM in `dependencyManagement` with `<scope>import</scope>` and `<type>pom</type>` and configure the compiler and Boot plugin yourself; it works, it is more to maintain, prefer the parent.

```xml
<dependencies>
  <!-- all groupId org.springframework.boot; no versions, the BOM supplies them -->
  <dependency><artifactId>spring-boot-starter-webmvc</artifactId>      ...</dependency>
  <dependency><artifactId>spring-boot-starter-validation</artifactId>  ...</dependency>
  <dependency><artifactId>spring-boot-starter-data-jpa</artifactId>    ...</dependency>
  <dependency><artifactId>spring-boot-starter-flyway</artifactId>      ...</dependency>
  <dependency><artifactId>spring-boot-starter-actuator</artifactId>    ...</dependency>
  <dependency><artifactId>spring-boot-starter-grpc-client</artifactId> ...</dependency>
  <dependency><artifactId>spring-boot-starter-test</artifactId><scope>test</scope>...</dependency>

  <dependency>
    <groupId>com.h2database</groupId><artifactId>h2</artifactId><scope>runtime</scope>
  </dependency>
</dependencies>
```

Six starters and a driver. `webmvc` brings Spring MVC, Tomcat 11, Jackson 3 and the web auto-configuration. `validation` brings Hibernate Validator 9.1 and Jakarta Validation 3.1 — note it is a *separate* starter, so if `@Valid` appears to do nothing, this is missing. `data-jpa` brings Hibernate 7.1, Jakarta Persistence 3.2, Spring Data JPA and HikariCP. `flyway` is section 2.2's lesson made concrete. `actuator` is Part 12; `grpc-client` is Part 9. The H2 driver is a plain library at `runtime` scope, which is correct — the datasource auto-configuration lives in modules the JPA starter already brought. What is *not* covered is the H2 web console, which has its own module; add the `h2` jar alone and you get a console that does not exist and does not say so.

```xml
<build>
  <finalName>molfar</finalName>
  <plugins>
    <plugin>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-maven-plugin</artifactId>
    </plugin>
  </plugins>
</build>
```

The plugin binds its `repackage` goal to `package`: Maven builds `target/molfar.jar` as usual, then the plugin rewrites it as an executable jar and renames the original to `molfar.jar.original`. `<finalName>molfar</finalName>` is what makes `just build-service`'s `cp service/target/molfar.jar bin/molfar.jar` work without a version in the path.

## 3.3 `@SpringBootApplication`, expanded

```java
package com.grazhda.molfar;

@SpringBootApplication
@EnableScheduling
public class MolfarApplication {
    public static void main(String[] args) {
        SpringApplication.run(MolfarApplication.class, args);
    }
}
```

Almost all of the behaviour is invisible. `@SpringBootApplication` composes three annotations.

**`@SpringBootConfiguration`** — a specialisation of `@Configuration` marking this class as a source of bean definitions. The specialisation exists so the test framework can find the application class by searching upward from a test's package, which is why moving `MolfarApplication` out of the root package breaks tests confusingly.

**`@ComponentScan`** — scan this class's package and everything beneath it for `@Component` and its stereotypes, registering each as a bean. The base package is implicit, and that is the important consequence: **`MolfarApplication` must live in the root package**, `com.grazhda.molfar`, with everything else beneath it. A class in a sibling package is invisible to the container, and — you will recognise the pattern — nothing tells you.

**`@EnableAutoConfiguration`** — run the auto-configuration classes contributed by classpath modules, each guarded by its conditions and notably by `@ConditionalOnMissingBean`. That condition is the mechanism behind Boot's central promise: auto-configuration backs off wherever you defined the bean yourself.

`@EnableScheduling` is explicit because molfar has scheduled jobs; without it, `@Scheduled` methods are parsed and never run. `SpringApplication.run` creates the context, drives scanning and auto-configuration, instantiates every singleton eagerly, starts Tomcat, then returns — while the container keeps a non-daemon thread alive so the JVM does not exit. It also registers the shutdown hook that makes graceful shutdown work.

## 3.4 Running it, and the build recipe

```bash
mvn spring-boot:run                    # development: fastest edit-run loop
mvn package && java -jar target/molfar.jar   # what actually ships
```

The executable jar is worth understanding once. Java cannot load nested jars, so Boot does not try to make it: `molfar.jar` holds your classes under `BOOT-INF/classes/`, your dependencies as untouched jars under `BOOT-INF/lib/`, and a launcher under `org/springframework/boot/loader/`. The manifest's `Main-Class` is that launcher, which installs a class loader able to read nested jars and then calls your `Start-Class`. Dependencies stay whole, so signatures survive and duplicate resources across jars do not silently collide the way they do in a shaded uber-jar. grazhda's Justfile wraps this:

```just
build-service:
    @mkdir -p bin
    cd service && mvn -q -B -DskipTests package
    cp service/target/molfar.jar bin/molfar.jar
```

One Boot 4.1 wrinkle applies directly. **`-DskipTests` no longer skips AOT processing.** It still skips running tests, but AOT work proceeds, so the recipe may be slower than expected and can fail on something unrelated to packaging. `-Dmaven.test.skip=true` skips test compilation and execution both. Confirm the effect on your setup with one timed build rather than assuming.

---

# Part 4 — Dependency injection, this project's way

## 4.1 Beans and the container

A **bean** is an object the container constructs and manages, knowing its type, dependencies, lifecycle and name. The default **scope** is singleton: one instance per context, created eagerly at start-up, shared by everything that asks. Assume singleton and design accordingly — **a singleton bean must be thread-safe or stateless**, because a `@RestController` handling two `zgard` requests at once is one object on two threads.

## 4.2 Component scanning and stereotypes

`@Component` is the base; `@Service`, `@Repository` and `@RestController` are specialisations. Functionally they are nearly identical, but they differ in intent and, in two cases, in behaviour. `@Service` marks an application-layer collaborator and adds nothing but documentation. `@Repository` marks a persistence adapter and *does* add behaviour: exception translation, converting vendor-specific `SQLException`s and JPA exceptions into Spring's `DataAccessException` hierarchy — worth using on hand-written persistence code so a constraint violation surfaces as `DataIntegrityViolationException`. `@Controller` / `@RestController` makes a class eligible for request mapping. `@Configuration` marks a source of `@Bean` methods. Use the specific one. It costs nothing and tells the next reader which layer they are in, which matters given Part 10's layering rule.

## 4.3 Constructor injection, and only constructor injection

```java
@Service
public class ChoreSubmissionService {

    private final ChoreJournal journal;
    private final Planner planner;
    private final MolfarProperties properties;

    public ChoreSubmissionService(ChoreJournal journal, Planner planner,
                                  MolfarProperties properties) {
        this.journal = journal; this.planner = planner; this.properties = properties;
    }

    public ChoreId submit(ChoreIntent intent, RepositoryScope scope, Budget budget) {
        var chore = Chore.submitted(intent, scope, budget, properties.maxStepsPerChore());
        journal.record(chore, planner.plan(chore));
        return chore.id();
    }
}
```

Since Spring 4.3 a class with a **single constructor** needs no `@Autowired` — the container uses it. So there is no Spring annotation in that constructor at all; it is an ordinary Java class. That is the point. With several constructors, exactly one must carry `@Autowired`; the better answer is usually one constructor.

**Resolution rules,** because you will meet them. The container matches each parameter by type. Exactly one bean of that type: done. None: start-up fails with `NoSuchBeanDefinitionException` — loudly, at boot, which is the good case. Several: start-up fails with `NoUniqueBeanDefinitionException`, and you disambiguate with `@Primary` on one, or `@Qualifier("dukhHealthSource")` on the parameter. Genuinely optional dependencies are declared `Optional<T>` or `@Nullable`.

## 4.4 `@Configuration` and `@Bean`, for types you do not own

Scanning only works on classes you can annotate. For a third-party type — or for a class in `core/` that must not import Spring, which is Part 10's whole point — write a factory method.

```java
@Configuration
public class CoreBeans {

    @Bean
    Planner planner(MolfarProperties properties) {
        return new Planner(properties.maxStepsPerChore(), properties.defaultBudget());
    }

    @Bean
    RunEngine runEngine(ChoreJournal j, StepExecutor e, RepositoryHealthSource h, Clock c) {
        return new RunEngine(j, e, h, c);
    }

    @Bean Clock systemClock() { return Clock.systemUTC(); }
}
```

The return type is the bean type, the parameters are injected as a constructor's would be, and the method name is the bean name. `Planner` and `RunEngine` are plain classes in `core/` with no framework annotation on them; this class is the seam where Spring learns how to build them.

`@Bean Clock systemClock()` looks trivial and is not: injecting a `Clock` rather than calling `Instant.now()` inside domain code is what makes time testable, and Part 11 depends on it. Note also that `@Configuration` classes are CGLIB-proxied by default so that calling `planner()` from inside `runEngine()` would return the same bean rather than a second instance; if you never do that — and you should not — `@Configuration(proxyBeanMethods = false)` skips the proxy.

## 4.5 Profiles

A profile is a named set of beans and properties, active or not. `@Profile("!test")` on a configuration class excludes it under `test`; `application-test.yaml` is layered over `application.yaml` for every active profile. Activate with `spring.profiles.active`, an environment variable, or `@ActiveProfiles("test")` on a test class. Use them sparingly. A service with `dev`, `test`, `local`, `laptop` and `ci` profiles has five configurations nobody exercises together, and the one that breaks is always the one you did not run. molfar needs `test` and arguably nothing else; everything that varies between engineers should be a property, not a profile.

## 4.6 Why field injection is forbidden here

```java
// Forbidden in this project.
@Service
public class ChoreSubmissionService {
    @Autowired private ChoreJournal journal;
    @Autowired private Planner planner;
}
```

It compiles, it runs, it is shorter, and it is banned for four reasons that compound.

**The object cannot be constructed correctly outside Spring.** `new ChoreSubmissionService()` gives you an instance with two null fields and no complaint. A plain JUnit test — the kind Part 11 wants most of your tests to be — cannot build a working instance without starting a context or reaching in with reflection. Constructor injection makes the test one line. This is not aesthetics: it is the difference between a suite that runs in 200 ms and one that runs in 40 seconds, and therefore the single largest lever on how much logic actually gets covered.

**The fields cannot be `final`.** Spring sets them after construction, so the object is mutable and there is no compiler-enforced guarantee that a collaborator is present and unchanged for its lifetime. With constructor injection every dependency is `final` and non-null from the first instruction of the first method.

**It hides design problems.** A constructor with nine parameters is ugly, and its ugliness is information: this class does too much. Nine `@Autowired` fields look tidy while carrying the same problem. Constructor injection makes coupling visible in the one place a reviewer always reads.

**Failure moves from start-up to runtime.** An unsatisfiable dependency fails at construction under constructor injection — the process refuses to start, which is what you want. Field injection can leave a partially populated object that fails later, on a request, with an NPE whose stack trace does not mention wiring. A fifth, smaller reason: constructor injection surfaces circular dependencies as a start-up failure rather than quietly permitting them (section 13.6). Setter injection is the third option and is not needed here.

Enforce the rule mechanically rather than in review. ArchUnit does it in a few lines (section 10.5), and a rule that runs in CI is worth more than a rule in a document.

---

# Part 5 — Configuration

## 5.1 `application.yaml`

Boot reads `application.properties` or `application.yaml` from `src/main/resources`, from a `config/` directory beside the jar, and from the working directory. YAML suits molfar because the tree is nested and grazhda's other configuration is YAML already.

```yaml
spring:
  datasource:
    url: jdbc:h2:file:${MOLFAR_HOME:${user.home}/.grazhda}/molfar;AUTO_SERVER=TRUE
  jpa: { open-in-view: false, hibernate: { ddl-auto: validate } }
  lifecycle: { timeout-per-shutdown-phase: 30s }

server: { port: 8420, shutdown: graceful }

management.endpoints.web.exposure.include: health,info,metrics,conditions,scheduledtasks

molfar:
  workspace-root: ${GRAZHDA_DIR:${user.home}/.grazhda}
  max-steps-per-chore: 40
  default-budget: { max-minutes: 30, max-repositories: 12 }
  dukh: { target: localhost:9090, call-timeout: 3s }
```

`${VAR:default}` is Spring's placeholder syntax, resolved against the whole environment, and it nests — `${MOLFAR_HOME:${user.home}/.grazhda}` reads an environment variable and falls back to a JVM system property. That is how one file serves every engineer's laptop without any of them editing it. Two lines are load-bearing. `spring.jpa.open-in-view: false` turns off a default that is wrong for this service (section 7.4). `ddl-auto: validate` means Hibernate checks the schema against the entities and refuses to start on a mismatch but never modifies it, because Flyway owns the schema (section 7.7).

## 5.2 Precedence

Boot merges many sources into one `Environment`. The part you will use, lowest priority first: defaults set in code; `application.yaml` inside the jar; `application-<profile>.yaml` inside the jar; `application.yaml` beside the jar or in `./config/`; the profile variant beside the jar; OS environment variables; JVM system properties; command-line arguments (`--molfar.max-steps-per-chore=10`); test property sources.

Higher wins. The consequence worth internalising: **anything in the file can be overridden from outside without rebuilding**, and environment variables map onto property names by upper-casing and replacing dots and dashes with underscores, so `molfar.max-steps-per-chore` is `MOLFAR_MAX_STEPS_PER_CHORE`. That relaxed binding is what makes it practical for the `grazhda` shell toolchain to start molfar with an override.

## 5.3 `@ConfigurationProperties` on records

You can inject a single value with `@Value("${molfar.max-steps-per-chore}")`. Do not build a service that way: `@Value` scatters string keys through the codebase, gives no type safety beyond conversion, and cannot be validated as a unit. Bind a whole tree to a record instead.

```java
@ConfigurationProperties(prefix = "molfar")
@Validated
public record MolfarProperties(
        @NotNull Path workspaceRoot,
        @Min(1) @Max(500) int maxStepsPerChore,
        @NotNull @Valid BudgetProperties defaultBudget,
        @NotNull @Valid DukhProperties dukh) {

    public record BudgetProperties(@Min(1) int maxMinutes, @Min(1) int maxRepositories) {}
    public record DukhProperties(@NotBlank String target, @NotNull Duration callTimeout) {}
}
```

Records work directly: Boot binds through the canonical constructor, so the result is immutable and every component injecting it gets a value that cannot change underneath it. Register with `@ConfigurationPropertiesScan` on the application class, or list types explicitly with `@EnableConfigurationProperties`. Then inject it like anything else.

Note the free conversions: `Path` from a string, `Duration` from `3s` or `PT3S`, `DataSize` from `10MB`, enums from their names, `List<T>` and `Map<String, T>` from YAML. `call-timeout: 3s` becomes a `Duration` with no code. Add `spring-boot-configuration-processor` as an optional dependency so your editor autocompletes and documents your own property names.

## 5.4 Fail at boot, not at the first request

`@Validated` on the record is the difference between two very different Tuesdays. Without it, `max-steps-per-chore: 0` starts cleanly and produces an empty plan at 02:00 during the nightly sweep, and you debug the planner. With it:

```
***************************
APPLICATION FAILED TO START
***************************

Property: molfar.max-steps-per-chore
Value: "0"
Reason: must be greater than or equal to 1
```

The process refuses to start and names the property, the value and the constraint. This requires `spring-boot-starter-validation` on the classpath — without it `@Validated` is inert and you are back to the first case, silently. Section 2.2's pattern again, and why that starter is in the pom even though molfar's REST surface is small. Constraints that are not annotations go in a compact component whose constructor throws:

```java
@Component
public class WorkspaceRootCheck {
    public WorkspaceRootCheck(MolfarProperties properties) {
        Path root = properties.workspaceRoot();
        if (!Files.isDirectory(root) || !Files.isWritable(root)) {
            throw new IllegalStateException(
                "molfar.workspace-root must be an existing writable directory: " + root);
        }
    }
}
```

Throwing from a constructor aborts context refresh and therefore start-up. A service running unattended for weeks should validate every assumption it can at second zero: a misconfigured workspace root failing at start-up costs a minute; the same value failing at 02:00 on day nine costs a morning.

## 5.5 The wrinkle: grazhda already has a `config.yaml`

grazhda's Go processes read `$GRAZHDA_DIR/config.yaml`, describing workspaces, projects and repositories with documented defaults and environment overrides. molfar needs some of that, because a chore's repository scope resolves against exactly those workspaces. So which file wins?

**Read `config.yaml` directly.** One source of truth, no drift. The cost is that molfar depends on a format specified by the Go side and needs its own parser plus its own reimplementation of the defaulting and override rules — precisely the code most likely to drift, silently and asymmetrically. Add a field with a default on the Go side and molfar's parser will not know.

**Duplicate the settings into `application.yaml`.** Idiomatic Spring, fully validated at boot, no second parser. The cost is two files that must agree with no mechanism to make them. Or **import it into Spring's environment** with `spring.config.import: optional:file:${GRAZHDA_DIR}/config.yaml`, which layers the imported keys into the `Environment` at defined precedence so binding and validation work normally and a missing file is non-fatal — elegant, and correct only if the file's *key structure* matches what molfar binds, since the binder ignores keys it does not recognise and a mismatch is therefore silent.

The recommendation: **treat `config.yaml` as data, not as configuration, and read it explicitly.** Workspace and repository definitions are input to the domain, the same category as the contents of a git repository — not settings for the process. So `molfar.workspace-root` is a Spring property, and the workspaces beneath it are loaded through a `core`-layer port whose adapter parses the file, unit-tested against fixture files copied from the Go side. That keeps ownership honest: the Go processes own the format, molfar reads it, and the reading is testable without a Spring context. Whichever route you take, write the decision down — the failure mode is two components disagreeing about which repositories exist.

---

# Part 6 — The web layer, contract-first

## 6.1 OpenAPI as the source of truth

`zgard` is Go, molfar is Java, and between them sits an HTTP API neither language can check. Either the Java defines the contract and you generate a Go client from a document produced by reflection, or you write the document by hand and generate from it on both sides. Contract-first is right here for the same reason grazhda uses protobuf for `dukh`: the contract becomes a reviewable artefact that changes deliberately. `contracts/molfar-api.yaml` sits beside `contracts/dukh.proto`, a breaking change shows up as a diff in a pull request, and both sides regenerate.

```yaml
paths:
  /api/chores:
    post:
      operationId: submitChore
      requestBody:
        content: { application/json: { schema: { $ref: '#/components/schemas/SubmitChoreRequest' } } }
      responses:
        '202': { description: accepted, planning started }
        '400':
          content: { application/problem+json: { schema: { $ref: '#/components/schemas/Problem' } } }
```

On the Java side, `openapi-generator-maven-plugin` with the `spring` generator and `interfaceOnly=true` produces an annotated interface per tag, which your controller implements. The compiler then enforces the contract: change the document, regenerate, and drift is a compile error.

Two cautions. Generator support for a given Spring generation lags the Spring release, and Boot 4's Jackson 3 packages are exactly what a generator template can get wrong — verify your generator emits Boot 4-compatible code before committing to it, and be prepared to hand-write the interfaces from the document if it does not. And if you go code-first with `springdoc-openapi` instead, check its Boot 4 compatibility too, since it hooks deeply into MVC internals.

## 6.2 `@RestController` and records

```java
@RestController
@RequestMapping("/api/chores")
public class ChoreController {

    private final ChoreSubmissionService submissions;
    private final ChoreQueryService queries;

    public ChoreController(ChoreSubmissionService submissions, ChoreQueryService queries) {
        this.submissions = submissions; this.queries = queries;
    }

    @PostMapping
    @ResponseStatus(HttpStatus.ACCEPTED)
    public ChoreAccepted submit(@Valid @RequestBody SubmitChoreRequest request) {
        ChoreId id = submissions.submit(
                new ChoreIntent(request.intent()),
                RepositoryScope.of(request.repositories()),
                request.budget().toDomain());
        return new ChoreAccepted(id.value(), "PLANNING");
    }

    @GetMapping("/{id}")
    public ChoreView get(@PathVariable UUID id) {
        return queries.findById(new ChoreId(id))
                .map(ChoreView::from)
                .orElseThrow(() -> new ChoreNotFoundException(id));
    }
}
```

`@RestController` is `@Controller` plus `@ResponseBody`, so every handler's return value is serialised to the body rather than resolved as a view name. Request and response bodies are **records**, in the web adapter package, and they are *not* domain types:

```java
public record SubmitChoreRequest(
        @NotBlank @Size(max = 2000) String intent,
        @NotEmpty List<@NotBlank String> repositories,
        @NotNull @Valid BudgetRequest budget) {

    public record BudgetRequest(@Min(1) @Max(480) int maxMinutes,
                                @Min(1) @Max(100) int maxRepositories) {
        public Budget toDomain() {
            return new Budget(Duration.ofMinutes(maxMinutes), maxRepositories);
        }
    }
}
```

Keeping these separate from `core`'s `Chore` costs a mapping method and buys the ability to change either side independently — which you will need the first time the domain grows a field that has no business being on the wire. Jackson 3 binds records through the canonical constructor with no extra annotation.

## 6.3 Validation

`@Valid` on the `@RequestBody` parameter runs Jakarta Bean Validation over the record, including nested records reached through a `@Valid` component. A violation throws `MethodArgumentNotValidException` before your handler body runs, becoming a `400` by default.

Two things to remember. **`@Valid` needs `spring-boot-starter-validation` on the classpath**, or it does nothing and every malformed request reaches your service layer. And validation at the edge is for *shape* — required, non-blank, in range, well-formed — while *rules* belong in `core`. "Intent must not be blank" is an edge concern; "a chore's budget may not exceed the workspace's remaining budget" is a domain rule, testable without an HTTP request. Do not let bean validation annotations creep into domain types; that is how framework dependencies get into `core`.

## 6.4 Errors: `@RestControllerAdvice` and `ProblemDetail`

Unhandled exceptions produce Boot's default error body, which is fine for a browser and poor for a CLI. RFC 9457 Problem Details is the standard shape, and Spring implements it as `ProblemDetail`.

```java
@RestControllerAdvice
public class MolfarExceptionHandler extends ResponseEntityExceptionHandler {

    @ExceptionHandler(ChoreNotFoundException.class)
    ProblemDetail handleNotFound(ChoreNotFoundException ex) {
        var problem = ProblemDetail.forStatusAndDetail(HttpStatus.NOT_FOUND, ex.getMessage());
        problem.setType(URI.create("https://grazhda.dev/problems/chore-not-found"));
        problem.setTitle("Chore not found");
        problem.setProperty("choreId", ex.choreId().toString());
        return problem;
    }

    @ExceptionHandler(BudgetExceededException.class)
    ProblemDetail handleBudget(BudgetExceededException ex) {
        var problem = ProblemDetail.forStatusAndDetail(HttpStatus.UNPROCESSABLE_ENTITY,
                "Requested budget exceeds the configured maximum");
        problem.setProperty("maxMinutes", ex.maximum().toMinutes());
        return problem;
    }
}
```

`@RestControllerAdvice` registers these across every controller. Extending `ResponseEntityExceptionHandler` gives `ProblemDetail` responses for Spring's own exceptions — validation failures, unreadable bodies, unsupported media types — without a handler each. The response carries `Content-Type: application/problem+json`, `setProperty` adds extension members, and `zgard` gets a machine-readable error rather than a string to parse. Set `spring.mvc.problemdetails.enabled: true` so the framework's default handling uses problem details too. The rule of thumb: **`core` throws domain exceptions that know nothing about HTTP; the advice translates them.** `BudgetExceededException` carries a requested budget and a maximum, not a status code. That is what lets the same exception be handled differently by the gRPC advice in Part 9.

## 6.5 API versioning

New in Boot 4: API versioning is auto-configured rather than hand-rolled.

```yaml
spring:
  mvc:
    apiversion:
      default: 1.0.0
      use:
        header: X-Molfar-Version
```

The same path can then be mapped once per version with `@GetMapping(path = "/{id}", version = "1.0.0")` and a second method at `version = "1.1.0"`. The `use` block chooses the strategy — a header here; query parameter, media-type parameter and path segment are the alternatives. `default` is what applies when a client sends no version, which is what keeps existing `zgard` builds working. `ApiVersionResolver`, `ApiVersionParser` and `ApiVersionDeprecationHandler` beans cover anything unusual, and two strategies at once need ordering declared in a `WebMvcConfigurer`'s `configureApiVersioning`.

Worth turning on early even at one version: `zgard` and molfar are built and installed independently, so an engineer will eventually run an old CLI against a new service, and a versioned contract turns that from a mysterious deserialisation failure into a legible negotiation. Verify the exact property spellings against your version's documentation — this area is new and sits close to `spring.mvc.*` neighbours that have existed for years.

---

# Part 7 — Persistence

## 7.1 What an ORM is actually doing

If you have never used one, here is the model, because everything else follows from it.

An ORM maps classes to tables and references to foreign keys. Its interesting part is not the mapping but the **persistence context** — a unit-of-work cache between your code and the database for the duration of a transaction. When you load an entity, the context keeps a reference to it and a snapshot of its state. When you change a field you write no SQL: at flush time the ORM compares object to snapshot and issues an `UPDATE` for what changed. Loading the same row twice in one transaction returns the *same object*. Objects the context tracks are **managed**; objects outside it are **detached**. That is the whole idea, and also the source of every surprise below. Writes happen at times you did not write down; reads may or may not hit the database; changing a field of a managed object is a database write, and forgetting that is how a debugging session accidentally persists something.

JPA (Jakarta Persistence 3.2) is the specification, Hibernate 7.1 the implementation, Spring Data JPA a layer above both that removes repository boilerplate.

## 7.2 Entities

```java
package com.grazhda.molfar.adapter.jpa;

@Entity
@Table(name = "chore")
public class ChoreEntity {

    @Id private UUID id;
    @Column(nullable = false, length = 2000) private String intent;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 32) private ChoreStatus status;

    @Column(name = "submitted_at", nullable = false) private Instant submittedAt;

    @OneToMany(mappedBy = "chore", cascade = CascadeType.ALL, orphanRemoval = true)
    @OrderBy("ordinal ASC")
    private List<StepEntity> steps = new ArrayList<>();

    protected ChoreEntity() { }   // required by JPA

    public ChoreEntity(UUID id, String intent, ChoreStatus status, Instant submittedAt) { ... }
}
```

Several deliberate choices.

**A `protected` no-argument constructor**, because JPA instantiates entities reflectively before populating them, and `protected` keeps it out of your API while remaining accessible to Hibernate's proxies.

**`@Enumerated(EnumType.STRING)`, always.** The default is `ORDINAL`, storing the enum's position as an integer. Insert a constant in the middle of `ChoreStatus` a month from now and every persisted row silently changes meaning. There is no case where the ordinal default is worth the risk.

**Assigned UUID identifiers** rather than a generated sequence. The domain creates a `ChoreId` before anything is persisted, so `core` never deals with an object whose identity appears only after a save — a substantial simplification for Part 10. The cost is index locality on a large table, irrelevant for a laptop journal. If you want database-generated identity, `@GeneratedValue(strategy = GenerationType.IDENTITY)` is the usual choice and forces an immediate `INSERT` at `persist` time.

**This is not a domain class.** It lives in `adapter/jpa`, it is mutable, it has a no-arg constructor and it knows about columns. `core`'s `Chore` is a different type with different rules.

**`equals` and `hashCode` are absent, intentionally** — section 13.5.

## 7.3 Identity and change tracking

Inside one transaction, `repository.findById(id)` called twice returns the same instance; the second call does not query. Across two transactions they are different instances for the same row, which is why `equals` on entities is delicate.

```java
@Transactional
public void markRunning(ChoreId id) {
    ChoreEntity chore = repository.findById(id.value()).orElseThrow();
    chore.setStatus(ChoreStatus.RUNNING);
    // no save() call: the UPDATE is issued at flush, before commit
}
```

There is no `save`. The entity is managed, the field changed, Hibernate writes it. Newcomers find this unsettling and are right to: a stray setter inside a read-only-looking method is a database write. It also means that to read an entity, modify it in memory and *not* persist it, you must detach it or map it to another type first.

## 7.4 Lazy loading, and the failure you will hit

`@OneToMany` associations are lazy by default: Hibernate gives you a proxy collection and fetches its contents on first access, but only while the persistence context is open.

```java
// No @Transactional here.
public ChoreView load(ChoreId id) {
    ChoreEntity chore = repository.findById(id.value()).orElseThrow();
    return new ChoreView(chore.getId(), chore.getSteps().size());   // boom
}
```

```
org.hibernate.LazyInitializationException:
  could not initialize proxy [ChoreEntity.steps] - no Session
```

`findById` opened a context for the duration of that call and closed it on return, so the entity came back detached and `steps` has nothing to fetch through.

Three fixes, in order of preference. **Widen the transaction** with `@Transactional(readOnly = true)` so the context is open while you map. **Fetch what you need in the query** — `@Query("select c from ChoreEntity c left join fetch c.steps where c.id = :id")` — which is better still because it is explicit and avoids N+1. Or **use an entity graph**, `@EntityGraph(attributePaths = "steps")`, on the repository method.

The fix you should *not* use is the one Boot has historically applied for you. `spring.jpa.open-in-view` defaults to `true`, keeping the persistence context open for the entire HTTP request and making the exception vanish. It also means lazy loads fire from inside serialisation, holding a database connection for the whole request, producing SQL with no visible call site, and turning a JSON serialisation of an entity graph into an unbounded cascade of queries. Boot logs a start-up warning about it for good reason. **Set it to `false` on day one** and treat the `LazyInitializationException`s it exposes as a map of wrong transaction boundaries — doing this after a hundred handlers depend on the behaviour is a much larger job.

## 7.5 Spring Data repositories

You do not write a DAO. You declare an interface and Spring Data implements it at start-up.

```java
public interface ChoreJpaRepository extends JpaRepository<ChoreEntity, UUID> {

    List<ChoreEntity> findByStatusOrderBySubmittedAtAsc(ChoreStatus status);

    List<ChoreEntity> findByStatusInAndSubmittedAtBefore(
            Collection<ChoreStatus> statuses, Instant cutoff);

    @Query("select c from ChoreEntity c left join fetch c.steps where c.id = :id")
    Optional<ChoreEntity> findByIdWithSteps(@Param("id") UUID id);
}
```

`JpaRepository` supplies `save`, `findById`, `findAll`, `delete`, `count`, paging and sorting. The others are **derived queries**: Spring Data parses the method name — `findBy`, property paths joined by `And`/`Or`, keywords like `In`, `Before`, `Between`, `IsNull`, `Containing`, an optional `OrderBy` — and generates the JPQL. A name it cannot parse, or one naming a property that does not exist, fails at start-up with a clear message. That is a good property: queries are validated at boot rather than on first use. Derived names stop paying around four conditions. `findByStatusAndSubmittedAtBetweenAndBudgetMaxMinutesLessThanOrderBySubmittedAtDesc` is not readable by anyone. Past that, write `@Query` with a text block, or a `Specification`, or drop to `JdbcClient` for a genuinely awkward report. Mixing styles in one repository is fine.

## 7.6 Transactions, and where the boundary belongs

`@Transactional` means: start a transaction if none is active, join the caller's if one is; commit on normal return; **roll back on an unchecked exception and *not* on a checked one** unless you say `rollbackFor`. That default surprises people every time. The boundary belongs at the **application service**, not on the repository and not on the controller.

```java
@Service
public class RunService {

    private final RunEngine engine;        // core: no Spring, no JPA
    private final ChoreJournal journal;    // core port, JPA adapter behind it

    public RunService(RunEngine engine, ChoreJournal journal) {
        this.engine = engine; this.journal = journal;
    }

    @Transactional
    public RunId startRun(ChoreId choreId) {
        Run run = engine.begin(journal.loadChore(choreId));
        journal.recordRunStarted(run);
        return run.id();
    }

    @Transactional(readOnly = true)
    public List<ChoreSummary> pending() {
        return journal.findByStatus(ChoreStatus.PENDING);
    }
}
```

One method, one unit of work: either the run row and its initial step rows all land or none do. On the repository it would be too fine-grained — two repository calls would be two transactions, and a crash between them leaves the journal inconsistent. On the controller it would be too coarse, holding a connection while Jackson serialises. `readOnly = true` is worth the keystrokes: Hibernate skips dirty-checking snapshots and the driver can optimise. And **never put a long-running operation inside a transaction**: executing a step means running `git` against a repository, which can take minutes, and a transaction open for minutes holds a pooled connection for minutes — with a default pool of ten, ten concurrent runs exhaust it and every HTTP request queues behind them. Structure it as short transaction to claim the step, no transaction while executing, short transaction to record the outcome.

**The trap: self-invocation.** `@Transactional` works by proxy — Spring wraps the bean in a proxy that opens a transaction before delegating. A call *through* the proxy is intercepted; a call from one method of the bean to another goes straight down `this` and bypasses it entirely.

```java
public void startAll(List<ChoreId> ids) {
    for (ChoreId id : ids) {
        startRun(id);          // no transaction: this.startRun, not proxy.startRun
    }
}

@Transactional
public RunId startRun(ChoreId id) { ... }
```

Every `startRun` above runs with no transaction. No warning, no log line, and the code reads correctly. The same applies to `@Async`, `@Cacheable`, `@Retryable` and every other proxy-based annotation. Move the annotated method to a different bean and inject it — usually the right design anyway, since the two methods sit at different levels of abstraction. `@Transactional` on a `private` or `final` method is the same failure in a different disguise: the proxy cannot override it, so nothing happens.

## 7.7 Flyway

Hibernate can generate your schema from your entities. Do not let it. `ddl-auto: update` is convenient in a tutorial and catastrophic in a journal that must survive weeks of restarts: it never drops or narrows anything, it makes different decisions on different Hibernate versions, and it leaves you with a schema nobody wrote down. Flyway owns the schema, as versioned SQL under `src/main/resources/db/migration`:

```sql
-- V1__create_chore.sql
CREATE TABLE chore (
    id UUID NOT NULL PRIMARY KEY, intent VARCHAR(2000) NOT NULL,
    status VARCHAR(32) NOT NULL, submitted_at TIMESTAMP NOT NULL);

CREATE TABLE step (
    id UUID NOT NULL PRIMARY KEY,
    chore_id UUID NOT NULL REFERENCES chore(id) ON DELETE CASCADE,
    ordinal INT NOT NULL, kind VARCHAR(64) NOT NULL, repository VARCHAR(512) NOT NULL,
    UNIQUE (chore_id, ordinal));

CREATE INDEX idx_chore_status ON chore(status, submitted_at);
```

The convention is `V<version>__<name>.sql` — capital `V`, a version, **two** underscores, a description. One underscore means the file is not recognised as a migration and is skipped without comment. Flyway records applied migrations in `flyway_schema_history` and refuses to start if a file that already ran has changed, which is the property that makes it trustworthy: **migrations are append-only; you never edit one that has shipped.** With Flyway owning the schema, `spring.jpa.hibernate.ddl-auto: validate` makes Hibernate check the entities against the real schema at start-up and fail loudly on a mismatch, catching the very common error of adding a field to an entity and forgetting the migration. And section 2.2 in context, because this is exactly where it bites: **`flyway-core` alone gives you a Flyway jar and no migrations.** You need `spring-boot-starter-flyway`. If molfar starts, logs nothing about migrations, and fails on a missing `chore` table, check `mvn dependency:tree` before you look at your SQL.

## 7.8 Choosing a database for a laptop service

The journal must survive restarts, `kill -9` and closing the lid, with no external infrastructure. Three real candidates.

**H2 in file mode** (`jdbc:h2:file:~/.grazhda/molfar`). Pure Java, no install, no daemon, first-class Hibernate dialect, real transactions, a web console for inspection, and one of the two embedded databases Boot still supports now Derby is deprecated. Weaknesses: less trodden as a durable store, so its failure modes are less well known; a file written by one H2 version may need migration to be read by the next, which for a long-lived journal is a genuine consideration; and only one JVM may open the file by default, though `AUTO_SERVER=TRUE` lets a second process connect through the first.

**SQLite.** The most battle-tested embedded database in existence, a stable single-file format, readable by a hundred tools — including the Go side of grazhda, which is not nothing. Weaknesses that matter here: Hibernate ships no official SQLite dialect, so you take a community one and its maintenance risk; the dynamic type system does not map cleanly onto JPA expectations; and one-writer-at-a-time concurrency means the scheduled sweep and an inbound `zgard` write can collide as `SQLITE_BUSY` unless WAL mode and a busy timeout are configured carefully. Excellent if you were writing JDBC by hand; awkward under an ORM.

**PostgreSQL in a container.** The most capable option, identical to production, exact type fidelity, real concurrency. It also requires Docker running for weeks, adds a process to start, stop and upgrade, and breaks the "no external infrastructure" requirement outright — every engineer who has not started the container gets a service that will not boot.

**Recommendation: H2 in file mode.** The deciding argument is not that H2 is the best database; it is that molfar is a single-process journal on one machine with one writer, small workload, trivial concurrency, and a hard requirement to start with `java -jar` and nothing else. H2 gives standard SQL, a proper dialect and honest transactions with zero setup. SQLite's dialect risk and writer-lock behaviour cost more under an ORM than its portability buys; Postgres costs a dependency the requirements exclude. Two things follow: keep the migration SQL portable, with standard types and no H2-specific syntax, so a future move to Postgres is a dialect change and not a rewrite; and keep one Testcontainers-based Postgres test (section 11.5), because "it worked on H2" and "the SQL is portable" are different claims.

---

# Part 8 — Scheduling and asynchronous work

## 8.1 `@Scheduled`

`@EnableScheduling` on the application class, then annotate methods on beans:

```java
@Component
public class DependencySweepJob {

    private final SweepService sweeps;

    public DependencySweepJob(SweepService sweeps) { this.sweeps = sweeps; }

    @Scheduled(cron = "${molfar.schedule.nightly-sweep}", zone = "Europe/Kyiv")
    public void nightlySweep() { sweeps.sweepAll(); }

    @Scheduled(cron = "${molfar.schedule.morning-report}", zone = "Europe/Kyiv")
    public void morningHealthReport() { sweeps.publishHealthReport(); }
}
```

A scheduled method takes no parameters and returns `void`. Three ways to say when: **`cron`**, a six-field expression `second minute hour day-of-month month day-of-week` — note the leading seconds field, so a five-field crontab expression pasted in is off by a factor you will notice, and always set `zone` or the schedule follows the JVM default and shifts under daylight saving; **`fixedDelay`**, waiting a period *after the previous run finishes*, which is self-limiting; **`fixedRate`**, starting every N units *from the previous start*, so a run longer than N leaves the next due immediately. Put the schedule in configuration rather than the annotation, as above, so it can change without a rebuild — and so the special value `-` can disable a job entirely, which is exactly what tests want.

## 8.2 The overlapping run, and what actually happens

Here is the trap, and it is subtler than "fixedRate overlaps".

By default Boot auto-configures a `ThreadPoolTaskScheduler` with **one thread**. That has a consequence people rarely think through: with one thread, scheduled tasks never overlap — but they also **block each other**. If the 02:00 sweep takes seven hours because a repository is enormous, the 08:30 health report does not run at 08:30. It runs at 09:00. Nothing warns you; the report simply appears late, or never. The obvious fix is to raise the pool:

```yaml
spring:
  task:
    scheduling:
      pool:
        size: 4
      thread-name-prefix: molfar-sched-
```

which unblocks the report and immediately reintroduces the real problem: with four threads, a cron job whose previous execution is still running **will** start a second one. Two concurrent dependency sweeps writing to the same journal is a data problem, not a performance problem. The same is true more sharply with virtual threads: `spring.threads.virtual.enabled: true` gets you a `SimpleAsyncTaskScheduler` with no pool and no bound, so overlap becomes the default rather than an edge case.

So the fix is not a scheduler setting. **The job must refuse to run concurrently with itself, in the job.** For a single-process service on one machine, an in-memory guard is enough and is honest about what it does:

```java
private final AtomicBoolean running = new AtomicBoolean(false);

@Scheduled(cron = "${molfar.schedule.nightly-sweep}", zone = "Europe/Kyiv")
public void nightlySweep() {
    if (!running.compareAndSet(false, true)) {
        log.warn("nightly sweep skipped: previous run still in progress");
        return;
    }
    try {
        sweeps.sweepAll();
    } finally {
        running.set(false);
    }
}
```

Three properties worth noting. The guard is `compareAndSet`, not `synchronized` — a second trigger should *skip*, not queue behind the first. The skip is logged at `warn`, so a chronically overrunning job is visible rather than invisible. And the reset is in a `finally`, because a guard that leaks on an exception disables the job permanently, which is worse than the problem it prevents.

If molfar ever runs as more than one process against one journal, an in-memory flag is not enough and you want a lock in the database — a row in a `job_lock` table claimed with a conditional update, or the ShedLock library, which does exactly this. Reach for it when the requirement appears, not before. And set `thread-name-prefix` regardless: in a thread dump of a service that has been up nine days, `molfar-sched-1` tells you something and `pool-3-thread-1` does not.

## 8.3 `@Async`

`@EnableAsync` plus `@Async` on a method makes the call return immediately and the body run on another thread. The method must return `void` or `CompletableFuture<T>` — a method returning a plain value with `@Async` on it returns `null` immediately, which is a very unpleasant bug. And **`@Async` is proxy-based**, so section 7.6's self-invocation trap applies identically: calling an `@Async` method from another method of the same bean runs it synchronously, silently.

Boot auto-configures an `AsyncTaskExecutor` when you have not defined one — a `ThreadPoolTaskExecutor` with eight core threads, or a virtual-thread `SimpleAsyncTaskExecutor` under `spring.threads.virtual.enabled`. Tune with `spring.task.execution.pool.*`, and choose the queue deliberately: bounded and full means rejected submissions, unbounded means they accumulate in memory, which over weeks is a slow leak.

**The 4.1 improvement worth knowing:** `@Async` methods now propagate context across the thread boundary automatically. In Boot 3 the trace ID and MDC entries were lost at the handoff, so the asynchronous half of an operation logged without a correlation ID — exactly where you most wanted one. Even so, use `@Async` sparingly here: executing a chore's steps is not fire-and-forget, since it must be recorded, resumable and interruptible, which wants an explicit executor and a journal row per step rather than an annotation. `@Async` is right for peripheral work — notifications, cache warming — and wrong for the core loop.

## 8.4 Graceful shutdown, and the interrupted run

Graceful shutdown is **enabled by default** in Boot 4 for Tomcat, Jetty and Reactor Netty. On `SIGTERM` the server stops accepting new requests at the network layer, in-flight requests finish, then the context closes. The grace period is `spring.lifecycle.timeout-per-shutdown-phase`, default 30 seconds; `server.shutdown: immediate` disables it. Note that your IDE's stop button may not send a real `SIGTERM`, so shutdown behaviour observed in an IDE is not evidence.

That handles HTTP. It does not handle molfar's actual problem: when the process stops, a chore run may be halfway through step four of nine, and **an interrupted run must be recorded as interrupted, not lost.** A run left in `RUNNING` forever is worse than one marked `INTERRUPTED`, because on restart nothing distinguishes it from a run genuinely in progress.

**Record the intent on the way down**, with a `SmartLifecycle` bean so you control ordering:

```java
@Component
public class RunSupervisor implements SmartLifecycle {

    private final Set<RunId> inFlight = ConcurrentHashMap.newKeySet();
    private volatile boolean running;
    // constructor injects ChoreJournal journal

    @Override public void start() { running = true; }
    @Override public boolean isRunning() { return running; }
    @Override public int getPhase() { return Integer.MAX_VALUE - 100; }

    @Override public void stop() {
        running = false;
        inFlight.forEach(id -> journal.markInterrupted(id, "molfar shut down mid-run"));
    }
}
```

`getPhase` matters: `SmartLifecycle` beans stop in descending phase order, so a high phase means this stops **before** lower-phase infrastructure. If the datasource has already closed, `markInterrupted` cannot write and you have a shutdown hook that logs an exception instead of doing its job. Fiddly enough to be worth testing rather than reasoning about.

**Reconcile on the way up**, because the shutdown path cannot be relied on — `kill -9`, a flat battery and a kernel panic all skip it. At start-up every run still marked `RUNNING` must be resolved, since none can legitimately be in progress in a process that has just started:

```java
@Component
public class OrphanedRunReconciler {
    public OrphanedRunReconciler(ChoreJournal journal) {
        int reclaimed = journal.markAllRunningAsInterrupted("molfar restarted");
        if (reclaimed > 0) log.warn("reclaimed {} interrupted run(s)", reclaimed);
    }
}
```

Doing the work in the constructor makes it part of context refresh, so it completes before the scheduler starts and before the first request is accepted. `ApplicationRunner` and `@EventListener(ApplicationReadyEvent.class)` both run *after* the web server is up, which for reconciliation is the wrong side of the line. Between the two halves the invariant holds: a run is `RUNNING` only while a live process is running it. That invariant is worth more than either mechanism individually, and the crash-safe half is the one that actually guarantees it.

---

# Part 9 — Talking to `dukh` over gRPC

## 9.1 Generating Java from the shared `.proto`

`contracts/dukh.proto` is the contract, already compiled into Go under `daemon/proto/`. The Java side generates from the same file — a copied contract is a contract that drifts. Boot 4.1 provides dependency management for `io.github.ascopes:protobuf-maven-plugin`, and selecting a gRPC starter on start.spring.io adds it. Proto files default to `src/main/proto`; for grazhda, point the plugin at the shared directory instead:

```xml
<plugin>
  <groupId>io.github.ascopes</groupId>
  <artifactId>protobuf-maven-plugin</artifactId>
  <configuration>
    <sourceDirectories>
      <sourceDirectory>${project.basedir}/../contracts</sourceDirectory>
    </sourceDirectories>
    <binaryMavenPlugins>
      <binaryMavenPlugin>
        <groupId>io.grpc</groupId><artifactId>protoc-gen-grpc-java</artifactId>
      </binaryMavenPlugin>
    </binaryMavenPlugins>
  </configuration>
</plugin>
```

Generate once and read the output: per service you get an abstract `<Service>Grpc.<Service>ImplBase` to extend on the server side, and `<Service>BlockingStub`, `<Service>Stub` and `<Service>FutureStub` on the client side. molfar is a client, so `DukhHealthGrpc.DukhHealthBlockingStub` is the type it cares about. Take the exact plugin configuration from a generated project rather than this snippet; coordinates and element names move, and a build that does not generate is at least a loud failure.

## 9.2 Auto-configuration and channels

Add `org.springframework.boot:spring-boot-starter-grpc-client`. (`spring-boot-starter-grpc-server` and `spring-boot-starter-grpc-server-test` are the server-side counterparts; molfar needs neither today.) The `spring-grpc` project provides the programming model underneath, but you never declare it — the starters bring it.

```yaml
spring:
  grpc:
    client:
      channel:
        dukh:
          target: static://localhost:9090
          keepalive:
            time: 30s
            timeout: 5s
          inbound:
            message:
              max-size: 8MB
```

**Note the singular `channel`.** Spring gRPC 1.0 used `spring.grpc.client.channels.<name>.address`; 1.1 — the version in Boot 4.1 — renamed it to `spring.grpc.client.channel.<name>.target`. A great deal of writing published between those releases uses the plural form, and if you use it the channel is simply not configured. The 1.1 migration guide lists the rest: `keep-alive-time` → `keepalive.time`, `max-inbound-message-size` → `inbound.message.max-size`, `idle-timeout` → `idle.timeout`, `default-deadline` → `default.deadline`. Check the names against your version's documentation rather than any blog post, this one included.

`target` accepts a scheme: `static://host:port` for a fixed address, `dns:///host:port` for DNS resolution. For a daemon on the same laptop, `static://localhost:9090`.

## 9.3 Importing and injecting a stub

Spring gRPC 1.0 scanned the classpath for generated stubs automatically. **1.1 does not**, and expects you to declare which you want:

```java
@SpringBootApplication
@EnableScheduling
@ConfigurationPropertiesScan
@ImportGrpcClients(target = "dukh", types = DukhHealthGrpc.DukhHealthBlockingStub.class)
public class MolfarApplication { ... }
```

`target` names the channel from `application.yaml` and `types` lists the stub classes to register as beans; the annotation also accepts `basePackages` / `basePackageClasses` to scan a package, `prefix` to prefix bean names, and `factory` to control how stubs are built. After that the stub is an ordinary bean, injected by constructor into an adapter that implements a `core` port:

```java
@Component
public class DukhRepositoryHealthSource implements RepositoryHealthSource {

    private final DukhHealthGrpc.DukhHealthBlockingStub dukh;
    private final Duration callTimeout;

    public DukhRepositoryHealthSource(DukhHealthGrpc.DukhHealthBlockingStub dukh,
                                      MolfarProperties properties) {
        this.dukh = dukh; this.callTimeout = properties.dukh().callTimeout();
    }
    ...
}
```

This class is the only place in molfar that knows gRPC exists.

A note on sources: some published examples show a different form, `@ImportGrpcClients(@GrpcClient(SomeGrpc.SomeBlockingStub.class))`, with nested annotations rather than attributes. The Spring gRPC 1.1 reference documentation shows the `target`/`types` form used above. Confirm against the version on your classpath — both spellings are in circulation and only one will compile.

## 9.4 Deadlines

**Set a deadline on every call.** A gRPC call with no deadline waits indefinitely, and a scheduled job blocked forever on a wedged daemon is a hang that looks like nothing at all.

```java
var response = dukh
        .withDeadlineAfter(callTimeout.toMillis(), TimeUnit.MILLISECONDS)
        .getRepositoryHealth(request);
```

`withDeadlineAfter` returns a *new* stub with the deadline attached and leaves the original unchanged, which is why calling it inline is safe on a shared singleton stub. Exceeding it throws `StatusRuntimeException` with status `DEADLINE_EXCEEDED`. Deadlines propagate: if `dukh` calls onward, the remaining budget travels with the call rather than restarting. This per-call deadline is separate from `spring.grpc.client.channel.dukh.default.deadline`, which sets a channel-wide default; setting both is reasonable.

## 9.5 When `dukh` is not running

This is the case that matters, because on a laptop it is the normal case. A blocking stub against an unreachable server throws `StatusRuntimeException` with status `UNAVAILABLE`. The failure modes to handle are `UNAVAILABLE` (not running, or the port is closed), `DEADLINE_EXCEEDED` (running but slow or wedged), and, less often, `UNIMPLEMENTED` — running but speaking an older contract, which is the interesting one because it means the shared `.proto` has drifted.

The design rule: **molfar must be useful without `dukh`.** Repository health is an input to planning, not a precondition for it, so the adapter degrades rather than propagating a gRPC exception into the domain:

```java
@Override
public RepositoryHealth healthOf(RepositoryRef repository) {
    try {
        ...
    } catch (StatusRuntimeException e) {
        Status.Code code = e.getStatus().getCode();
        if (code == Status.Code.UNAVAILABLE || code == Status.Code.DEADLINE_EXCEEDED) {
            log.debug("dukh unavailable ({}), treating {} health as unknown", code, repository);
            return RepositoryHealth.unknown(repository);
        }
        throw new RepositoryHealthUnavailableException(repository, e);
    }
}
```

`RepositoryHealth.unknown(...)` is a `core` concept and the planner is written to handle it: an unknown-health repository is included in a plan but flagged in the proposal, so the human approving it can see health was not checked. That is a domain decision made in domain code; the adapter's only job is to translate a transport failure into it. Log unavailability at `debug`, not `warn`. A daemon down for six hours would otherwise produce thousands of identical warnings and train you to ignore the log. Surface the condition through the Actuator health endpoint (section 12.1) instead, where a monitor sees it once rather than ten thousand times.

If molfar ever exposes a gRPC server, `@GrpcAdvice` with `@GrpcExceptionHandler` methods is the centralised translation layer, exactly parallel to `@RestControllerAdvice` — a handler returns a `Status` such as `Status.NOT_FOUND.withDescription(...)`. Be careful with imports: there are two `@GrpcExceptionHandler` types, and the one in the advice package is the annotation you want; the other is a functional interface and using it produces a handler that never fires.

---

# Part 10 — Keeping the domain out of Spring

This is the most important section in the guide. Everything before it is mechanics; this is the decision that determines whether molfar is pleasant to work on in a year.

## 10.1 The rule

> **The domain logic lives in a plain-Java `core` package with no framework dependencies, unit-tested with plain JUnit and no Spring context. Spring is the outer layer: persistence, scheduling, HTTP, wiring.**

No `@Component`, no `@Service`, no `@Transactional`, no `@Entity`, no `@JsonProperty`, no `jakarta.*`, no `org.springframework.*` anywhere under `core/`. The only imports are `java.*` and other `core` types. This mirrors what the Go side already does — `libs/` holds the domain, `cli/` and `daemon/` hold the adapters — and the reasons are the same in both languages.

## 10.2 Why

**Because tests stop being slow.** A test of `Planner` needing no context starts in milliseconds. A test that starts a context costs one to three seconds, and Spring caches contexts by *configuration*, so every distinct combination of profiles, property overrides and mock beans builds a new one. The difference between a two-second suite and a two-minute one is the difference between running tests on every save and running them before a commit, and that changes how much of the logic actually gets covered. This is the single largest lever on test coverage in a Spring project.

**Because the interesting logic becomes readable.** Planning a chore is a pure function over values: intent, scope and budget in, an ordered list of steps out. As plain Java it is a class you read top to bottom. As a `@Service` with three `@Autowired` collaborators, two of them repositories, it is a class you must run to understand.

**Because framework upgrades stop being domain events.** Boot 3 to 4 moved auto-configuration packages, renamed starters and replaced Jackson 2 with Jackson 3 including its package names. Every one of those lands in the adapter layer. If `Chore` carried `@Entity` and `@JsonProperty`, that upgrade would touch your domain model; because it does not, `core` compiles unchanged.

**Because the domain gets the right shape.** JPA entities need a no-arg constructor and mutable fields; Jackson needs accessible properties. Both push you toward anaemic, mutable, get-and-set objects. Freed of them, `Chore` can be immutable, validate its invariants in its constructor, be a record or a sealed interface, and refuse to exist in an invalid state. `ChoreEntity` absorbs the framework's requirements so `Chore` does not have to. And **because the boundary is where the design lives**: writing the ports forces you to name what the domain actually needs from outside, and "I need to record that a run started" is smaller and clearer than "I need a `JpaRepository<ChoreEntity, UUID>`".

The cost is real: **you write mapping code**, `Chore` ↔ `ChoreEntity` and `Chore` ↔ `ChoreView`. That is genuine repetitive work, and on a small CRUD service it would not pay for itself. molfar is not a small CRUD service — it is a planner and an execution engine with a durable journal, where the logic is the point and the storage is incidental. Here the trade is clearly worth it. Be honest that it is a trade.

## 10.3 Ports and adapters

The domain declares interfaces for what it needs; adapters implement them.

```java
// core/port/ChoreJournal.java — no imports outside java.* and core.*
public interface ChoreJournal {
    void record(Chore chore, Plan plan);
    Chore loadChore(ChoreId id);
    void recordRunStarted(Run run);
    void recordStepOutcome(RunId runId, int ordinal, StepOutcome outcome);
    void markInterrupted(RunId runId, String reason);
    int markAllRunningAsInterrupted(String reason);
    List<ChoreSummary> findByStatus(ChoreStatus status);
}
```

Note what it does *not* say: nothing about SQL, tables, entities, transactions, `Optional<ChoreEntity>` or `Pageable`. It is expressed entirely in domain terms and would be equally implementable over a file, an in-memory map, or Postgres. That is the test of a good port: **if you cannot implement it in memory in twenty lines, it is leaking the database's shape into the domain.** The adapter is where Spring reappears:

```java
@Repository
public class JpaChoreJournal implements ChoreJournal {

    private final ChoreJpaRepository chores;

    public JpaChoreJournal(ChoreJpaRepository chores) { this.chores = chores; }

    @Override
    public void record(Chore chore, Plan plan) {
        ChoreEntity entity = ChoreMapper.toEntity(chore);
        plan.steps().forEach(step -> entity.addStep(StepMapper.toEntity(step)));
        chores.save(entity);
    }

    @Override
    public Chore loadChore(ChoreId id) {
        return chores.findByIdWithSteps(id.value())
                .map(ChoreMapper::toDomain)
                .orElseThrow(() -> new ChoreNotFoundException(id.value()));
    }
}
```

`ChoreEntity` never escapes this package; `Chore` never enters it except as an argument to be mapped. The dependency direction is strictly inward: `adapter` knows about `core`, `core` knows nothing about `adapter`. This also answers "how does the core persist anything without importing JPA": it does not persist anything. It calls `journal.record(chore, plan)` on an interface it owns, and something on the other side deals with rows. In a unit test, that something is twenty lines of `HashMap`.

## 10.4 The package layout

```
service/src/main/java/com/grazhda/molfar/
├── MolfarApplication.java        @SpringBootApplication — the root package
│
├── core/                         NO framework imports, anywhere below here
│   ├── chore/    Chore, ChoreId, ChoreIntent, ChoreStatus, Budget, RepositoryScope
│   ├── plan/     Plan, Step, StepKind, Planner
│   ├── run/      Run, RunId, RunEngine, StepOutcome
│   ├── proposal/ Proposal, ProposalId, Approval, Diff
│   ├── health/   RepositoryHealth, HealthStatus
│   └── port/     ChoreJournal, StepExecutor, RepositoryHealthSource, WorkspaceCatalog
│
├── adapter/
│   ├── web/      ChoreController, SubmitChoreRequest, ChoreView, MolfarExceptionHandler
│   ├── jpa/      ChoreEntity, StepEntity, RunEntity, ChoreJpaRepository, JpaChoreJournal, mappers
│   ├── grpc/     DukhRepositoryHealthSource
│   ├── git/      ProcessStepExecutor
│   └── yaml/     GrazhdaConfigWorkspaceCatalog
│
├── application/  ChoreSubmissionService, RunService, SweepService   @Service, @Transactional
├── scheduling/   DependencySweepJob, HealthReportJob, RunSupervisor
└── config/       MolfarProperties, CoreBeans, WorkspaceRootCheck
```

`MolfarApplication` sits in the root package because `@ComponentScan` starts from there.

`application/` is thin and should stay thin: transaction boundaries and orchestration — load, call `core`, save — and nothing else. When a method there starts containing an `if` about domain rules, that `if` belongs in `core`. This layer is where the two common leaks happen: business logic drifting outward because it was convenient, and JPA entities drifting inward because mapping felt like a chore. Both are worth catching in review, and better still in a test.

## 10.5 Enforcing it

A rule that lives only in a document decays. ArchUnit makes it a test:

```java
class LayeringTest {

    private static final JavaClasses classes =
            new ClassFileImporter().importPackages("com.grazhda.molfar");

    @Test void coreDependsOnNothingButJavaAndItself() {
        noClasses().that().resideInAPackage("..core..")
                .should().dependOnClassesThat()
                .resideInAnyPackage("org.springframework..", "jakarta..", "tools.jackson..",
                                    "com.fasterxml..", "io.grpc..", "org.hibernate..")
                .check(classes);
    }

    @Test void noFieldInjection() {
        noFields().should().beAnnotatedWith(Autowired.class).check(classes);
    }
}
```

Two tests, no Spring context, sub-second, failing in CI the moment somebody adds `@Entity` to a domain class at the end of a long day. That is worth more than any amount of documentation.

---

# Part 11 — Testing

## 11.1 The right pyramid

The instinct with Spring is to write `@SpringBootTest` for everything, because it works and requires no thought about what a test actually needs. Resist it, for one measurable reason: a full context costs one to three seconds, and the framework caches contexts by configuration, so forty test classes with slightly different setups can spend minutes doing nothing but building contexts.

The shape molfar wants: **most** tests plain JUnit over `core/` with no Spring at all; **some** slices — `@DataJpaTest`, `@WebMvcTest`, `@JsonTest` — over the adapters; **a handful** of `@SpringBootTest` for wiring, start-up and one or two end-to-end paths. Nearly all the interesting assertions live in the bottom layer, which is only possible because of Part 10.

## 11.2 Plain JUnit for `core`

```java
class PlannerTest {

    private final Planner planner = new Planner(40, new Budget(Duration.ofMinutes(30), 12));

    @Test
    void ordersDependencyUpdatesBeforeVerification() {
        var chore = Chore.submitted(new ChoreIntent("bump jackson across the backend"),
                RepositoryScope.of(List.of("api", "worker")),
                new Budget(Duration.ofMinutes(20), 5), 40);

        assertThat(planner.plan(chore).steps()).extracting(Step::kind)
                .containsExactly(StepKind.UPDATE_DEPENDENCY, StepKind.UPDATE_DEPENDENCY,
                                 StepKind.BUILD, StepKind.TEST);
    }

    @Test
    void refusesAPlanThatExceedsTheBudget() {
        var chore = Chore.submitted(new ChoreIntent("upgrade everything"),
                RepositoryScope.of(twentyRepositories()),
                new Budget(Duration.ofMinutes(5), 2), 40);

        assertThatThrownBy(() -> planner.plan(chore))
                .isInstanceOf(BudgetExceededException.class)
                .hasMessageContaining("2 repositories");
    }
}
```

No annotations beyond `@Test`, no context, under a millisecond. The whole planning ruleset fits in a file of these and the suite stays fast enough to run on every save.

Where a `core` class needs a collaborator, provide a hand-written fake — an `InMemoryChoreJournal` backed by a `LinkedHashMap` and a `List<StepOutcome>` is twenty lines, reusable across the package, and it lets you assert on *what was recorded* rather than *which methods were called*, which is a considerably better test. Mockito is available and appropriate for adapter tests; for the domain, a fake that actually behaves is clearer.

## 11.3 Slices

A slice starts a partial context containing only the beans one layer needs.

**`@DataJpaTest`** — JPA, Hibernate, a `DataSource` and your repositories; no web layer, no `@Service` beans. It is transactional and rolls back after each test. By default it replaces your datasource with an embedded one, which for molfar-on-H2 is what you were using anyway; `@AutoConfigureTestDatabase(replace = NONE)` keeps the real one.

```java
@DataJpaTest
class ChoreJpaRepositoryTest {

    @Autowired private ChoreJpaRepository chores;

    @Test
    void findsPendingChoresOldestFirst() {
        chores.save(chore("later",   Instant.parse("2026-09-02T10:00:00Z")));
        chores.save(chore("earlier", Instant.parse("2026-09-01T10:00:00Z")));

        assertThat(chores.findByStatusOrderBySubmittedAtAsc(ChoreStatus.PENDING))
                .extracting(ChoreEntity::getIntent).containsExactly("earlier", "later");
    }
}
```

This is where derived queries and Flyway migrations get verified: the migrations run against the test database, so a migration that does not match the entities fails here rather than at start-up on somebody's laptop.

**`@WebMvcTest(ChoreController.class)`** — MVC infrastructure, JSON conversion, validation and your `@RestControllerAdvice`; no JPA, no datasource, no services. Collaborators come as mocks via `@MockitoBean` (the Boot 3.4+ replacement for the deprecated `@MockBean`).

```java
@WebMvcTest(ChoreController.class)
class ChoreControllerTest {

    @Autowired private MockMvc mvc;
    @MockitoBean private ChoreSubmissionService submissions;
    @MockitoBean private ChoreQueryService queries;

    @Test
    void rejectsABlankIntentWithProblemDetails() throws Exception {
        mvc.perform(post("/api/chores").contentType(MediaType.APPLICATION_JSON)
                .content("""
                         {"intent":"", "repositories":["api"],
                          "budget":{"maxMinutes":10,"maxRepositories":2}}"""))
           .andExpect(status().isBadRequest())
           .andExpect(content().contentTypeCompatibleWith("application/problem+json"));
    }
}
```

This exercises validation and error translation — the things most likely to be wrong — with no database anywhere near it.

**`@JsonTest`** — serialisation only. Small, fast, and the right place to pin the wire format `zgard` depends on, which given Jackson 3's changed defaults (section 2.5) is not theoretical.

If a slice appears to do nothing, check for the corresponding test starter such as `spring-boot-starter-webmvc-test`. Section 2.2 wearing a test-scoped hat.

## 11.4 `@SpringBootTest` and `RestTestClient`

Reserve the full context for what genuinely needs it: that the application starts at all, that the beans wire together, and one or two end-to-end paths.

```java
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@AutoConfigureRestTestClient
@ActiveProfiles("test")
class MolfarSmokeTest {

    @Autowired private RestTestClient client;

    @Test
    void submittingAChoreReturnsAnIdentifier() {
        client.post().uri("/api/chores").contentType(MediaType.APPLICATION_JSON)
              .body("""
                    {"intent":"bump jackson", "repositories":["api"],
                     "budget":{"maxMinutes":10,"maxRepositories":2}}""")
              .exchange().expectStatus().isAccepted()
              .expectBody().jsonPath("$.choreId").exists();
    }
}
```

`RestTestClient` is Boot 4's fluent test client, replacing `TestRestTemplate`. With `RANDOM_PORT` it makes real HTTP calls to the started server; with `@AutoConfigureMockMvc` and no server it runs over `MockMvc`, faster and still exercising the whole MVC stack. **The gotcha: HTTP test clients are no longer auto-configured in Boot 4** — without `@AutoConfigureRestTestClient` the injection fails, and it reads like a wiring mystery rather than a missing annotation. Two habits keep full-context tests cheap: use the *same* configuration across all of them so one cached context is reused, and use `@MockitoBean` sparingly, since each distinct set of mocked beans is a distinct context.

## 11.5 Testcontainers

Testcontainers starts a real database in Docker for the duration of a test, and Boot integrates through `@ServiceConnection`:

```java
@SpringBootTest
@Testcontainers
@Tag("docker")
class PostgresPortabilityTest {

    @Container @ServiceConnection
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:17");

    @Test
    void flywayMigrationsApplyCleanlyOnPostgres() { ... }
}
```

`@ServiceConnection` replaces the old `@DynamicPropertySource` boilerplate: Boot reads the container's JDBC URL, user and password and points the datasource at it.

Be honest about the requirement: **Testcontainers needs a working Docker daemon.** Where Docker is not running these tests fail, not skip. Given that molfar deliberately has no external infrastructure, keep a small number of them behind a `@Tag("docker")` excluded from the default `just test-java` and run by a separate recipe. Their value is narrow but real: they are the only thing that can tell you whether the SQL in your migrations is actually portable, which section 7.8 said you would want to know before any future move off H2.

## 11.6 AssertJ

`spring-boot-starter-test` brings AssertJ, and it is the assertion library to use — fluent, discoverable through autocompletion, and much better than JUnit's built-ins for collections and exceptions.

```java
assertThat(plan.steps()).hasSize(4)
        .extracting(Step::repository).containsExactlyInAnyOrder("api", "worker");

assertThat(proposal).usingRecursiveComparison()
        .ignoringFields("createdAt").isEqualTo(expected);
```

`usingRecursiveComparison` is particularly useful with immutable records: it compares field by field without requiring you to write `equals` on every domain type just to make a test work. Add ArchUnit as a test dependency for section 10.5's rules — three architecture tests running in under a second are the cheapest insurance in the project.

---

# Part 12 — Running it for weeks

## 12.1 Actuator: health and metrics

`spring-boot-starter-actuator` adds management endpoints; by default only `health` is exposed over HTTP, so expose the rest explicitly with `management.endpoints.web.exposure.include`. molfar binds to localhost on a laptop with nothing in front of it, so exposing `loggers` — which changes log levels at runtime — is fine there and would not be on a public interface.

The endpoints that earn their place: **`health`**, is it alive and is the journal reachable; **`metrics`**, where `/actuator/metrics/hikaricp.connections.active` is the gauge that tells you whether section 7.6's advice about long transactions is being followed; **`scheduledtasks`**, listing every `@Scheduled` method with its schedule and next execution, which is the fastest way to answer "why did the nightly sweep not run"; **`loggers`**, raising a package to `DEBUG` on a running process and lowering it again without a restart, which on a service up for nine days is worth a great deal; and **`conditions`**, section 2.3's diagnostic.

Write a health indicator for what is actually specific to this service:

```java
@Component
public class DukhHealthIndicator implements HealthIndicator {

    private final RepositoryHealthSource health;

    public DukhHealthIndicator(RepositoryHealthSource health) { this.health = health; }

    @Override public Health health() {
        return health.isReachable()
                ? Health.up().withDetail("target", health.target()).build()
                : Health.status("DEGRADED")
                        .withDetail("target", health.target())
                        .withDetail("effect", "repository health reported as unknown")
                        .build();
    }
}
```

Note the status: `DEGRADED`, not `DOWN`. molfar without `dukh` is diminished, not broken (section 9.5), and a health endpoint reporting `DOWN` for a condition the service is designed to tolerate is one people stop believing. Custom statuses need `management.endpoint.health.status.http-mapping` if you want a non-200 response. Register domain metrics too — a gauge for `molfar.runs.active` costs three lines and answers a question the JVM metrics cannot.

## 12.2 Logging and rotation

Boot uses Logback by default (Log4j2 by swapping starters) and logs to the console. For a process running for weeks you want a file, and one that does not grow without bound.

```yaml
logging:
  file:
    name: ${MOLFAR_HOME:${user.home}/.grazhda}/logs/molfar.log
  logback:
    rollingpolicy:
      file-name-pattern: ${LOG_FILE}.%d{yyyy-MM-dd}.%i.gz
      max-file-size: 10MB
      max-history: 14
      total-size-cap: 500MB
  level:
    root: INFO
    com.grazhda.molfar: DEBUG
```

`total-size-cap` matters most on a laptop: it is the hard ceiling on disk consumed by logs, and without it `max-history` alone can still let a chatty week fill a disk. Fourteen days of history matches how long the process is expected to run.

If you switch to Log4j2, 4.1 adds rotation *strategies* under `logging.log4j2.rollingpolicy.strategy`: `size` (default), `time`, `size-and-time`, and `cron` (configured with `logging.log4j2.rollingpolicy.cron`). Time-based rotation is often more useful for a long-running local process, because it makes "what happened on Tuesday" a single file. For structured output, `logging.structured.format.file: ecs` (or `logstash`, or `gelf`) writes JSON instead of text. Whether that is an improvement depends entirely on whether anything will parse it: if the only reader is an engineer with `less`, plain text with a good pattern is better.

Two habits from the start. Log identifiers, not prose — `log.info("run {} step {} completed in {}ms", runId, ordinal, elapsed)` is greppable and "Step finished successfully" is not. And use parameterised logging rather than concatenation, so the message is not built when the level is disabled.

## 12.3 Graceful shutdown, configured

```yaml
server:
  shutdown: graceful
spring:
  lifecycle:
    timeout-per-shutdown-phase: 30s
```

It is on by default in Boot 4; setting it explicitly documents intent. The timeout is the budget for each shutdown phase — long enough for in-flight requests to finish and for section 8.4's `RunSupervisor` to write its interruption records, short enough that a stop script does not appear to hang. If your `stop()` needs longer than thirty seconds it is doing too much, and the crash-safe start-up reconciliation is what you should be relying on instead.

## 12.4 Memory, and the flags worth setting

Be realistic. Untuned, a service with webmvc, JPA, Hibernate, Flyway, a gRPC client and Actuator settles at roughly 100–200 MB of heap in use with a default maximum of one quarter of physical RAM — which on a 32 GB laptop means the JVM will let the heap grow to 8 GB before collecting seriously; 80–130 MB of metaspace, high because of Spring's proxy generation and Hibernate's metamodel; another 80–150 MB of code cache, thread stacks, GC structures and direct buffers; and a resident set of 250–400 MB in total.

The most valuable flag is a heap cap, because the default is far larger than this service needs and an uncapped heap on a laptop competes with the IDE:

```bash
java -Xms128m -Xmx384m \
     -XX:MaxMetaspaceSize=192m \
     -XX:+UseSerialGC \
     -XX:+ExitOnOutOfMemoryError \
     -jar bin/molfar.jar
```

Setting `-Xms` and `-Xmx` equal avoids resize pauses at the cost of always holding the memory. `-XX:MaxMetaspaceSize` turns a metaspace leak — the classic symptom of repeated context refreshes — into a fast failure rather than slow machine-wide degradation. `-XX:+UseSerialGC` is deliberate: G1 is the default and is better for large heaps and many cores, but it carries per-region overhead and background threads that a single-user service with a small heap does not benefit from; Serial GC on a 384 MB heap has short pauses and a smaller footprint. `-XX:+ExitOnOutOfMemoryError` is the difference between a process that dies and gets restarted and one that limps along throwing `OutOfMemoryError` from random threads for a day. In a container, prefer `-XX:MaxRAMPercentage=70` to a fixed `-Xmx`, since the JVM reads the cgroup limit. Two flags to know and not use in production: `-XX:TieredStopAtLevel=1` cuts start-up by disabling C2, which is good for tests and bad for a process that runs for weeks and wants its hot paths optimised; and AppCDS class-data sharing genuinely helps start-up but only pays back if you restart often, which molfar does not.

On virtual threads: with Java 25, `spring.threads.virtual.enabled: true` gives a virtual thread per request instead of a pool and removes the thread-pool sizing question for the web layer. Java 25 has also resolved the worst pinning problem — `synchronized` blocks no longer pin a carrier thread — which was the main reason for caution under Java 21. Reasonable for molfar's HTTP layer, but re-read section 8.2 first: it also switches the scheduler to an unbounded `SimpleAsyncTaskScheduler`, making concurrent execution of a scheduled job the default rather than the exception.

## 12.5 Starting and stopping it from another process

```bash
nohup java -Xmx384m -jar "$GRAZHDA_DIR/bin/molfar.jar" \
      --spring.config.additional-location="file:$GRAZHDA_DIR/molfar.yaml" \
      > "$GRAZHDA_DIR/logs/molfar.out" 2>&1 &
echo $! > "$GRAZHDA_DIR/molfar.pid"

until curl -sf http://localhost:8420/actuator/health > /dev/null; do sleep 0.2; done

kill -TERM "$(cat "$GRAZHDA_DIR/molfar.pid")"
```

Three points. Use `SIGTERM`, never `SIGKILL` — `SIGTERM` runs the shutdown hook and therefore section 8.4's interruption recording. Poll `/actuator/health` for readiness rather than sleeping a fixed interval, because start-up time varies with disk cache and migration work. And `--spring.config.additional-location` layers an external file *on top of* the bundled `application.yaml`, which is almost always what you want; `--spring.config.location` replaces the lot and will surprise you. For something more robust than a pidfile, a systemd user unit with `Type=simple`, `Restart=on-failure` and `TimeoutStopSec=45` gives supervision, restart-on-crash and log integration on Linux, with `launchd` the macOS equivalent. Either is a better answer than a shell script the day the laptop reboots.

## 12.6 OpenTelemetry

Boot 4 ships `org.springframework.boot:spring-boot-starter-opentelemetry` for tracing, metrics and log export without requiring Actuator, configured under `management.opentelemetry.tracing.export.otlp.endpoint` and friends, with sampling under `management.tracing.sampling.probability`.

For molfar on a laptop with no collector, the honest answer is that you probably do not need trace export — one process, one user, and the log file plus Actuator metrics answer most questions. What you *do* want is the trace and span identifiers in the log lines, so a chore's HTTP submission, its planning, its scheduled execution and its gRPC call to `dukh` share a correlation ID you can grep for. That works without a collector, and Boot 4.1's automatic `@Async` context propagation (section 2.6) is what makes it survive the thread boundary that previously broke the chain. `management.opentelemetry.enabled: false` disables the SDK while keeping propagators active — the right setting for a laptop, and one line to flip if you ever point the service at a collector.

---

# Part 13 — Gotchas and a cheat sheet

## 13.1 The starter is the contract

Section 2.2 at length, repeated because it costs the most. Adding a raw library in Boot 4 gives you the library with no Spring wiring and no error message. `flyway-core` without `spring-boot-starter-flyway` runs no migrations and says nothing. Whenever a feature seems to be doing nothing at all, before debugging your code: `mvn dependency:tree`, then `debug=true`, then `/actuator/conditions`.

## 13.2 Jackson 3

Core classes moved to `tools.jackson.*`; annotations stayed at `com.fasterxml.jackson.annotation`. `ObjectMapper` gives way to `JsonMapper` with an immutable builder. `JacksonException` is unchecked. `JsonSerializer` is `ValueSerializer`. Dates default to ISO-8601 strings instead of numeric timestamps, which is a wire-format change. Pin your contract with a `@JsonTest`.

## 13.3 Proxy-based annotations do nothing on self-invocation

```java
public void processAll(List<ChoreId> ids) {
    ids.forEach(this::processOne);      // no transaction, no async, no cache
}

@Transactional
public void processOne(ChoreId id) { ... }
```

`@Transactional`, `@Async`, `@Cacheable`, `@Retryable`, `@PreAuthorize` all work through a proxy wrapping the bean, and a call through `this` never touches it. Same for `private`, `final` and `static` methods, which the proxy cannot intercept. There is no warning. Move the annotated method to another bean and inject it.

## 13.4 `LazyInitializationException`

Touching a lazy association outside the transaction that loaded it fails. Fix by widening the transaction, by fetching with `join fetch` or `@EntityGraph`, or by mapping to a DTO inside the transaction. Do **not** fix it by leaving `spring.jpa.open-in-view` at its `true` default — that hides the problem behind unbounded query cascades during serialisation and a connection held for the whole request. Set it to `false` on day one.

## 13.5 `equals` and `hashCode` on JPA entities

An entity's identifier is often null before persistence and non-null after, so an `equals` based on the id changes behaviour mid-lifecycle: put such an entity in a `HashSet` before saving and its hash changes after the save, making it unfindable in the set it is in.

The safe rules: with a database-generated id, either do not override `equals`/`hashCode` at all — identity semantics, which the persistence context already guarantees within a transaction — or use the id for `equals` and a **constant** `hashCode`, which is legal, degrades hash lookups to a linear scan within a bucket, and is fine for the small collections an entity graph holds. Never base `equals` on mutable business fields. molfar sidesteps most of this by assigning UUIDs in the domain before persistence, so the id is never null. Even so, keep `equals` off `ChoreEntity` and let `Chore` — an immutable record — carry value semantics.

## 13.6 Constructor injection cycles

Two beans that require each other cannot both be constructed first, so start-up fails:

```
The dependencies of some of the beans in the application context form a cycle:
   choreSubmissionService → runService → choreSubmissionService
```

That is the rule doing its job. The temptation is `spring.main.allow-circular-references=true`, which makes the message go away by falling back to setter injection and leaves two objects each holding a partially initialised reference to the other. Do not. A cycle means the responsibilities are wrong: extract the shared behaviour into a third bean, or invert one direction with an event. `@Lazy` on one side also breaks the cycle by injecting a proxy, and is a legitimate escape hatch that should still make you uncomfortable.

## 13.7 Configuration properties that bind to nothing

Relaxed binding means `maxStepsPerChore` in YAML binds to `max-steps-per-chore`. The reverse mistake does not fail either: a property named `molfar.max-step-per-chore` (singular `step`) binds to nothing, the record field keeps its default of `0`, and nothing is reported. Likewise a `@ConfigurationProperties` class never registered — no `@ConfigurationPropertiesScan`, no `@EnableConfigurationProperties` — gives a bean that cannot be injected or, worse, one with every field at its default. Three defences: `@Validated` with `@Min`/`@NotNull` on every field so an unbound value fails at boot rather than defaulting; `spring-boot-configuration-processor`, so your editor autocompletes your own property names; and a test asserting that a known property actually arrived.

## 13.8 Smaller ones, quickly

- **The application class must be in the root package.** `@ComponentScan` starts from its package; a sibling package is never scanned.
- **Cron in `@Scheduled` has six fields**, starting with seconds. A five-field crontab expression means something else entirely.
- **Always set `zone` on a cron schedule**, or it follows the JVM default and moves under daylight saving.
- **`@Transactional` rolls back on unchecked exceptions only.** A checked exception commits.
- **`@Enumerated` defaults to `ORDINAL`.** Always write `EnumType.STRING`.
- **`@Async` on a method returning a plain value returns `null`.** Return `void` or `CompletableFuture<T>`.
- **`@MockBean` is deprecated**; use `@MockitoBean`.
- **HTTP test clients are not auto-configured in Boot 4.** `@AutoConfigureRestTestClient` is required.
- **`-DskipTests` no longer skips AOT processing in 4.1.** Use `-Dmaven.test.skip`.
- **`spring.grpc.client.channel` is singular** in Spring gRPC 1.1; the plural `channels` was 1.0 and binds to nothing now.
- **Flyway migration filenames need two underscores**: `V1__name.sql`. One underscore is silently ignored.
- **`@Value` on a `final` field does not work** — field injection cannot write to a final field, and the value ends up null.

## 13.9 One-page reference

**Annotations you will actually use**

| Annotation | What it does |
|---|---|
| `@SpringBootApplication` | `@SpringBootConfiguration` + `@EnableAutoConfiguration` + `@ComponentScan` |
| `@Component` / `@Service` / `@Repository` | Register a scanned bean; `@Repository` adds exception translation |
| `@RestController` | `@Controller` + `@ResponseBody` |
| `@Configuration` / `@Bean` | Factory methods for types you cannot annotate |
| `@ConfigurationProperties` | Bind a property subtree to a record |
| `@ConfigurationPropertiesScan` | Find every such record in scanned packages |
| `@Validated` | Run bean validation on a properties class at boot |
| `@Profile` | Include this bean only under a named profile |
| `@Transactional` | Transaction boundary; proxy-based; rolls back on unchecked only |
| `@Scheduled` | `cron` (six fields, set `zone`), `fixedDelay`, `fixedRate` |
| `@EnableScheduling` / `@EnableAsync` | Turn the above on; without them the annotations are inert |
| `@Async` | Run on another thread; `void` or `CompletableFuture` only |
| `@RestControllerAdvice` / `@ExceptionHandler` | Translate exceptions into `ProblemDetail` |
| `@Valid` / `@RequestBody` / `@PathVariable` | Web-layer binding and validation |
| `@Entity` / `@Id` / `@Column` | JPA mapping; adapter layer only |
| `@Enumerated(EnumType.STRING)` | Never accept the ordinal default |
| `@ImportGrpcClients` | Register generated gRPC stubs as beans |
| `@ImportHttpServices` / `@HttpExchange` | Declarative HTTP clients from an interface |
| `@GrpcAdvice` / `@GrpcExceptionHandler` | Centralised gRPC error translation |

**Starters for molfar** (all `org.springframework.boot`)

| Artifact | Brings |
|---|---|
| `spring-boot-starter-webmvc` | Spring MVC, Tomcat 11, Jackson 3 (was `-web` in Boot 3) |
| `spring-boot-starter-validation` | Hibernate Validator 9.1; required for `@Valid` to do anything |
| `spring-boot-starter-data-jpa` | Hibernate 7.1, Spring Data JPA, HikariCP |
| `spring-boot-starter-flyway` | Flyway **and** its auto-configuration module |
| `spring-boot-starter-actuator` | health, metrics, conditions, loggers, scheduledtasks |
| `spring-boot-starter-grpc-client` | Spring gRPC client and channel configuration |
| `spring-boot-starter-opentelemetry` | OTLP tracing, metrics and log export |
| `spring-boot-starter-test` | JUnit Jupiter, AssertJ, Mockito, test context |
| `spring-boot-starter-webmvc-test` | Web test slice support |

**Properties worth memorising**

| Property | Why |
|---|---|
| `spring.jpa.open-in-view: false` | Turn off the wrong default on day one |
| `spring.jpa.hibernate.ddl-auto: validate` | Flyway owns the schema; Hibernate checks it |
| `server.shutdown: graceful` | Default in Boot 4; state it anyway |
| `spring.lifecycle.timeout-per-shutdown-phase` | Budget for in-flight work at shutdown |
| `spring.task.scheduling.pool.size` | Default is 1: jobs block each other |
| `spring.threads.virtual.enabled` | Java 21+; also changes the scheduler's overlap behaviour |
| `management.endpoints.web.exposure.include` | Only `health` is exposed by default |
| `logging.logback.rollingpolicy.total-size-cap` | Hard ceiling on disk used by logs |
| `spring.grpc.client.channel.<name>.target` | Singular `channel` in Spring gRPC 1.1 |
| `spring.mvc.apiversion.default` | Old `zgard` builds against a new molfar |
| `spring.mvc.problemdetails.enabled: true` | RFC 9457 bodies from the default handler |
| `debug: true` | The condition evaluation report |

**When something silently does nothing, in order**

1. `mvn dependency:tree` — is the starter there, or only the library?
2. `debug=true` — did the auto-configuration match, not match, or never appear?
3. `/actuator/conditions` — the same report, on a running process.
4. Is the class inside `MolfarApplication`'s package tree?
5. Is the call going through the proxy, or through `this`?
6. Is the enabling annotation present — `@EnableScheduling`, `@EnableAsync`, `@ConfigurationPropertiesScan`?

---

## Where to go next

Read the Spring Boot 4.1 release notes and the 4.0 migration guide end to end. They are short, they are the authority on everything in Part 2, and they will inoculate you against the very large volume of Boot 3 material you will otherwise find first.

Then read one auto-configuration class — `FlywayAutoConfiguration` in `spring-boot-flyway` is a good one, being small and now easy to locate. Seeing the `@ConditionalOnClass`, `@ConditionalOnMissingBean` and `@AutoConfiguration(after = ...)` annotations that decide whether your Flyway gets configured turns section 2.2 from a rule you memorised into a mechanism you understand, and it is the fastest way to stop finding auto-configuration mysterious.

Then build `core/plan/Planner` first, with plain JUnit and no `pom.xml` changes at all. It takes a chore and produces an ordered plan within a budget, it is the most interesting logic in the service, and building it before any Spring exists is the most direct way to prove Part 10's rule to yourself.
