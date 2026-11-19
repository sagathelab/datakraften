# Datakraften 3.0 – veikart

**Status:** Produktretning og plan, ikke en beskrivelse av funksjonalitet som allerede er levert.  
**Utgangspunkt:** `v0.12.23`, 11. november 2026.  
**Tidligere planer:** [2.0-veikartet](datakraften_2.0_roadmap.md) og [plattformplanen](datakraften-platform-plan.md) er historiske og inneholder både leverte og uleverte forslag. Dette dokumentet er retningen for arbeidet mot 3.0.

## 1. Visjon

> A better starting point for your development machine.  
> Datakraften brings your runtimes, shell, editors and AI tools together in a configuration you can understand and share.

Dette er fortsatt kjernen. Datakraften skal i tillegg være en **støttespiller gjennom hele utviklermiljøets livsløp**: fra å oppdage og forstå verktøy, via installasjon og oppsett, til prosjektarbeid, diagnostikk og trygt vedlikehold.

Målet er å bli en foretrukket løsning for utviklere fordi produktet er nyttig i hverdagen, forståelig, pålitelig og verdt tilliten. Det er en ambisjon som må vinnes gjennom faktiske brukerbehov og kvalitet – ikke noe en versjonsnummerering i seg selv oppnår.

### Produktbeslutning: Én sammenhengende løsning for maskin og prosjekter

Datakraften skal både hjelpe utvikleren med **den lokale utviklermaskinen** og tilby **prosjektfunksjonalitet som en sentral del av produktet**. `dk` skal gjøre det enkelt å finne, sette opp og vedlikeholde utviklerverktøy, runtimes og rammeverk, og deretter bruke riktige versjoner, miljøvariabler og kommandoer i hvert prosjekt. Node er den første konkrete versjonsstyrte arbeidsflyten, ikke en begrensning på den langsiktige retningen.

Repoet er førstevalget for delt prosjektoppsett: samme utvikler og team kan arbeide i flere repoer med ulike behov, uten å skifte global maskinkonfigurasjon. For arbeidsstasjonen kan `dk init` importere en ekstern YAML-fil som flettes med standardene og lagres lokalt; brukeren eier og kan endre resultatet. Hvert repo kan i tillegg ha en egen `datakraften.yaml` for prosjektmiljøet. Dette er besluttet produktretning, ikke funksjoner som allerede er implementert.

```text
Personlig utviklermaskin: shell, editor, Git og personlige verktøy
  ├── Repo A: Node 22, prosjektvariabler og egne kommandoer
  ├── Repo B: Node 24, andre variabler og egne kommandoer
  └── Repo C: andre runtimes og rammeverk etter behov
```

Prosjektmiljøet avgrenses i første omgang til prosessen som `dk` starter og dens barn. Det er ikke en sikkerhetssandkasse og isolerer ikke automatisk systempakker, tjenester, filer eller nettverk. Et prosjekt skal ikke endre et annet prosjekts oppsett eller maskinens globale standard som en skjult bivirkning.

### Produktløftet for 3.0

1. **Oppdag:** Finn relevante utviklerverktøy og forstå hva de gjør, hvem de passer for og hvilke alternativer som finnes.
2. **Sett opp:** Bygg en utviklermaskin fra en forståelig konfigurasjon, uten å miste kontroll over eksisterende oppsett.
3. **Arbeid:** Kjør prosjekter med riktige verktøyversjoner, miljøvariabler og oppgaver på en forutsigbar måte.
4. **Forstå:** Se hva som faktisk er installert, aktivt, administrert, mangler eller trenger oppmerksomhet.
5. **Vedlikehold:** Gjennomgå og gjennomfør oppdateringer med tydelig informasjon om kilde, konsekvenser og eventuelle feil.
6. **Beskytt:** Få handlingsrettet sikkerhetsinformasjon om de verktøyene og installasjonsveiene Datakraften faktisk kan kontrollere.
7. **Del:** Gjør oppsett og prosjektkrav forståelige og delbare i et team, uten å legge hemmeligheter i konfigurasjonen.

### Produktprinsipper

- **Forståelig først:** Forklar hva som skal skje, hvorfor og hva brukeren kan gjøre videre. God diagnostikk er en del av produktet.
- **Brukeren eier maskinen:** Respekter eksisterende verktøy, konfigurasjon og preferanser. Ikke overta, avinstaller eller overskriv uten tydelig valg.
- **Sikkert som standard:** Skille mellom å lese konfigurasjon og å kjøre den. Minst mulig privilegier, kjente kilder, synlige risikoer og ingen skjulte bakgrunnsendringer.
- **Samarbeid med eksisterende verktøy:** Bruk modne pakkebehandlere og versjonsbehandlere når det gir mening. Ikke bygg en generell pakkebehandler for å duplisere dem.
- **Ærlige resultater:** Ukjent er ikke det samme som sikkert, installert er ikke det samme som tilgjengelig, og en sjekksum alene beviser ikke opphav.
- **Lokal kontroll:** Ingen hemmelig innsamling av data. Nettverkstilgang, lagring av tilstand og eventuelle fremtidige skytjenester skal være forståelige og dokumenterte.
- **Delbarhet uten tvang:** Et godt standardoppsett, men plass til individuelle valg og eksisterende team- og prosjektkonvensjoner.

## 2. Hvem skal vi hjelpe?

| Bruker | Behov | Ønsket resultat |
| --- | --- | --- |
| Utvikleren som starter på nytt | Oppdage aktuelle verktøy og sette opp maskinen | Et fungerende oppsett med tydelige valg og neste steg |
| Utvikleren med flere prosjekter | Håndtere ulike versjoner og kommandoer | Riktig prosjektmiljø uten manuell omkonfigurering av maskinen |
| Utvikleren som bruker AI-verktøy | Holde orden på CLI-er, editorintegrasjoner og oppdateringer | Se hva som brukes, hvor det kommer fra og hvordan det vedlikeholdes |
| Konsulenten | Bytte mellom team og prosjekter | Forstå og kjøre hvert miljø uten å ødelegge et annet |
| Team- eller plattformansvarlig | Dele anbefalte oppsett og gjøre onboarding enklere | En etterprøvbar konfigurasjon med kontrollert endring og klare unntak |
| Sikkerhetsbevisst utvikler | Redusere risiko i lokale verktøy og leverandørkjeder | Finne konkrete problemer, forstå usikkerheten og velge en handling |

## 3. Faktisk utgangspunkt og viktige begrensninger

Dette veikartet tar utgangspunkt i implementasjonen, ikke i eldre løfter i README eller planer.

### Finnes i dag

- `dk init`, `dk apply`, `dk doctor`, `dk status`, `dk update`, `dk upgrade` og `dk npm setup` finnes. `dk apply` har bekreftelse og forhåndsvisning; støtten for flagg og JSON varierer mellom kommandoene og må dokumenteres og verifiseres per kommando.
- Konfigurasjonskildene er `default`, `custom` og `team`. Teamkonfigurasjon hentes fra en URL. Faktisk validering, tillit og feilhåndtering må vurderes før dette omtales som en sikker teamløsning.
- Oppsettet bruker eksisterende verktøy, blant annet systemets pakkebehandler, Homebrew, `fnm` og `uv`. AI-verktøy og editorer er del av produktets oppsett og deteksjon, med ulik støtte avhengig av plattform.
- `dk apply` setter i dag opp Fish. Det finnes generatorer for Bash og Zsh, men de blir ikke generelt aktivert av `dk apply`. Fish-oppsettet inneholder `fnm --use-on-cd`; det betyr ikke at automatisk Node-bytte er aktivt i alle eksisterende shell eller miljøer.
- `dk doctor` viser informasjon om miljøet, blant annet kjente Homebrew- og Go-installasjoner utenfor `PATH`. Den er ikke en generell sikkerhetsskanner.
- `dk update` har i dag mål for Homebrew, fnm, uv og utvalgte npm-verktøy. Homebrew-oppdateringen kan omfatte alle Homebrew-pakker; den skal ikke fremstilles som begrenset til Datakraften-eide pakker.
- Utgivelsesflyten bygger Linux og macOS for amd64 og arm64, med SHA256-filer og installasjonsskript. En sjekksum hentet fra samme kilde som binæren er integritetskontroll, ikke en uavhengig bekreftelse av utgiver.

### Finnes ikke som komplett 3.0-løsning

- Prosjektbasert `dk run`, prosjektstyrte versjoner og miljøvariabler er ikke implementert.
- Verktøykatalog med sammenligninger, oppdaterte rangeringer og dokumentert målemetode er et mål, ikke en ferdig tjeneste.
- Kontinuerlig sikkerhetsovervåking, full sårbarhetsdekning, signaturkontroll for alle kilder og garantert sikre oppdateringer finnes ikke.
- Eldre beskrivelser av profiler, pakkevalg, plattformstøtte og flagg må ikke behandles som fasit uten kontroll mot kode og tester.

## 4. Avgrensning og samspill med andre løsninger

Datakraften skal forene veiledning, oversikt og konkrete handlinger. Det kan bruke en pakkebehandler til installasjon og en versjonsbehandler til prosjektmiljøer. Brukeren skal kunne forstå hvilken komponent som gjør hva.

[Mise](https://mise.jdx.dev/) tilbyr blant annet verktøyversjoner, miljøvariabler, oppgaver, søkbart verktøyregister og maskinoppsett. Datakraften konkurrerer derfor om noen av de samme behovene. Vår mulige forskjell er en samlet, tilgjengelig opplevelse av **oppdagelse, valg, arbeidsstasjon, prosjekt og vedlikehold**, med særlig oppmerksomhet på sikkerhet og AI-verktøy. Dette må valideres med brukere.

**Ikke mål for 3.0:** egen universell pakkebehandler, antivirus/EDR, fullstendig beskyttelse mot ukjente angrep, en ny hemmelighetsforvalter, erstatning for CI eller en garanti for at alle maskiner blir identiske. Native Windows-støtte, en skyplattform og bedriftsdashboard er heller ikke forpliktelser for 3.0.

## 5. Produktområde: Oppdag og forstå verktøy

### Mål

Datakraften.no skal være et sted utviklere kan finne aktuelle verktøy, forstå bruksområdene og ta egne valg. CLI-en skal kunne knytte valg til dokumentasjon og installasjonsmuligheter der støtte faktisk finnes.

### Leveranser

- En redaksjonelt vedlikeholdt katalog med tydelige kategorier, blant annet AI-verktøy, språk og runtimes, shell og terminal, editorer, containere og sikkerhet.
- For hvert verktøy: beskrivelse, bruksområde, offisiell kilde, plattformstøtte, lisens/prismodell der informasjonen kan kontrolleres, installasjonsmetoder og lenke til relevant dokumentasjon. Oppgi når informasjonen sist ble kontrollert.
- Vis om `dk` faktisk kan installere, oppdatere eller bare beskrive verktøyet på den aktuelle plattformen. Skil mellom testet støtte og ikke verifisert støtte.
- Søk, kategorier og sammenligning av relevante alternativer. Begynn med et begrenset, kvalitetssikret utvalg i stedet for å love en komplett katalog.
- Oppgi hvilke installasjonsvalg som er anbefalt, hvorfor de er valgt og hvilke rettigheter de krever når dette er kjent.
- Utforsk en CLI-inngang, for eksempel `dk search` og `dk info`. Navn og grensesnitt er forslag og må valideres før implementering.

### Popularitet og «topplister»

- Skil mellom **populært**, **trender**, **nylig oppdatert** og **redaksjonelt anbefalt**. Dette er ulike påstander.
- Oppgi datakilde, måleperiode, tidspunkt for siste oppdatering, utvalgsmetode og kjente begrensninger. Vis «ukjent» når data ikke finnes eller er gamle.
- Stjerner, nedlastinger og aktivitet må ikke presenteres som bevis på antall brukere, kvalitet eller sikkerhet. Ikke slå sammen ulike målinger til en skjult totalscore.
- Redaksjonelle anbefalinger skal ha en begrunnelse og ikke forveksles med betalt plassering. Eventuell sponsing må merkes tydelig.
- Avklar API-rettigheter, kvoter, vedlikeholdsansvar og hvordan feil rettes før automatiske lister publiseres.

### Ferdig når

En ny bruker kan finne og sammenligne verktøy i minst to prioriterte kategorier, se kilde og oppdateringsdato, og skille et anbefalt verktøy fra et verktøy som bare er populært. Installasjonsstøtten som vises, samsvarer med testet funksjonalitet.

## 6. Produktområde: Arbeidsstasjon og trygg tilstand

### Mål

`dk init`, `dk apply`, `dk status`, `dk doctor` og `dk update` skal gi en sammenhengende opplevelse fra første oppsett til løpende bruk.

### Leveranser

- Gjennomgå eksisterende konfigurasjon og implementasjon. Dokumenter hva som faktisk støttes, og rett avvik mellom CLI, nettside, README og tester.
- Et tydelig skille mellom ønsket tilstand, oppdaget tilstand og hva Datakraften selv administrerer. Oppgi kilde, aktiv versjon, plassering, eier og usikkerhet når dette kan fastslås.
- Pålitelig oppsett og diagnostikk for valgte shell: start med Fish og Bash, og utvid til Zsh etter dokumentert testing. Ikke endre standardshell uten uttrykkelig valg.
- Identifiser konflikter mellom systeminstallert Node, `fnm`, `nvm`, mise og andre relevante verktøy. Forklar hvilken kjørbar fil som vinner i gjeldende `PATH`, uten å endre eierskap automatisk.
- Vis en konkret plan før endringer, begrens endringer til det brukeren valgte, og oppgi delvis feil og neste steg. `--dry-run` må være reell for kommandoene som tilbyr den.
- Bevar brukerinnhold utenfor merkede blokker. Endringer skal være idempotente og håndtere avbrudd og uventede eksisterende filer på en trygg måte.
- Forhåndsvis oppdateringer og omfang, spesielt når en underliggende pakkebehandler oppdaterer mer enn verktøy Datakraften administrerer. Ikke lov rollback der den underliggende installasjonsmetoden ikke støtter det.

### Ferdig når

På testede WSL-, Fedora- og Ubuntu-oppsett kan en bruker forhåndsvise, anvende, gjenta og diagnostisere oppsettet uten uventet overtakelse av eksisterende verktøy. `status`, `doctor`, JSON og synlige resultater beskriver samme tilstand. macOS-krav fastsettes i plattformmatrisen før lansering.

## 7. Produktområde: Prosjektmiljø og `dk run`

### Mål

Prosjektfunksjonalitet er en prioritert del av 3.0: Et repo skal kunne beskrive runtimes og versjoner, støttede verktøy og rammeverk, ikke-hemmelige miljøvariabler og navngitte oppgaver. En utvikler skal enkelt kunne oppdage hva prosjektet trenger, få hjelp til å sette det opp og kjøre det uten å endre maskinens globale standardversjoner. Teamet deler kravene i repoet; hver utvikler kan ha flere prosjekter med forskjellige krav på samme maskin.

En runtime som Node og et prosjektrammeverk er ulike ting. Første leveranse skal styre Node-versjonen og kunne kjøre prosjektets eksisterende kommandoer, for eksempel et rammeverks `dev`-skript. Støtte for å installere eller opprette bestemte rammeverk skal komme via eksplisitte, dokumenterte integrasjoner; det er ikke en påstand om at 3.0 automatisk administrerer alle rammeverk eller prosjektavhengigheter.

Følgende er **et illustrerende forslag**, ikke en eksisterende kommando eller vedtatt endelig syntaks:

```yaml
# datakraften.yaml i prosjektroten
version: 1
project:
  tools:
    node: "22"
    python: "3.14"
  env:
    NODE_ENV: development
    PORT: "3000"
  tasks:
    dev: npm run dev
    test: npm test
```

```bash
# Foreslått brukeropplevelse
cd mitt-prosjekt
dk project status  # viser krav, valgt versjon, kilde og mangler
dk project setup   # viser og gjennomfører valgt prosjektoppsett
dk run dev
dk run test
```

Kommandoene og YAML-strukturen er illustrasjoner av ønsket resultat, ikke eksisterende funksjoner eller en låst spesifikasjon. `status` skal være lesende; et `setup`-steg skal forklare kilder og endringer og innhente nødvendige valg før installasjon. `dk run` skal forklare manglende krav og kunne kjøre en navngitt oppgave i prosjektets miljø etter at innholdet er godkjent.

### Konkrete produktkrav

- **Oppdage:** Fra repoet eller en undermappe skal brukeren kunne se prosjektrot, kilde til krav, ønsket og aktiv versjon, tilgjengelig backend og mangler uten at inspeksjon kjører prosjektkode eller installerer noe.
- **Sette opp:** Brukeren skal kunne klargjøre prosjektet med et tydelig valg, kilde og forhåndsvisning av nødvendige installasjoner. Eksisterende kompatible installasjoner skal gjenbrukes når det er trygt og mulig.
- **Versjonsstyre:** To repoer skal kunne velge forskjellige Node-versjoner side om side. `dk run` skal bruke repoets valgte versjon også når en annen Node står først i brukerens vanlige `PATH`.
- **Konfigurere miljø:** Prosjektet kan angi vanlige variabler og krav til lokalt tilgjengelige hemmeligheter. Verdier som er hemmelige skal ikke kreves i delte filer eller vises i logger. Prosessmiljøet skal ikke endre forelder-shellet.
- **Kjøre:** Prosjektet kan gi navn til oppgaver som `dev`, `build` og `test`; brukeren kan liste og kjøre dem i riktig arbeidsmappe med prosjektets verktøy og variabler.
- **Samarbeide:** Konfigurasjonen kan versjonskontrolleres i repoet, gjennomgås i pull requests og brukes av teamet uten en egen, maskinglobal teamkonfigurasjon. Lokal tilpasning og CI-bruk krever dokumenterte regler for tillit og overstyring.
- **Feilsøke:** Ved ukjent prosjekt, manglende runtime, uklar versjon, backendkonflikt, feilende installasjon eller avslått tillit skal `dk` forklare årsaken og en mulig neste handling.

### Kontrakt som må utformes og testes

- **Fil og søk:** Prosjektfilen heter `datakraften.yaml`. `dk` skal oppdage den i gjeldende mappe; i et Git-repo søkes det også oppover til og med repoets rot, med nærmeste fil først. Ikke fortsett til et annet repo eller over repoets rot. Utenfor et repo brukes filen i gjeldende mappe; eksplisitt valg av fil kan vurderes separat. Dokumenter håndtering av undermapper, nestede repoer, worktrees og symboliske lenker, og test at prosjektfiler aldri forveksles med `~/.config/datakraften/config.yaml`.
- **Oppdagelse kontra handling:** `dk` skal ved hver kommandoavvikling kunne finne ut om prosjektfilen finnes i den aktuelle konteksten. Dette innebærer ikke at alle kommandoer leser innstillinger som påvirker dem, kjører oppgaver, installerer eller endrer maskinen. Prosjektbevisste kommandoer skal bruke filen; maskinkommandoer som `dk init` og `dk apply` fortsetter å arbeide med den lokale maskinkonfigurasjonen. Ingen automatisk kjøring eller installasjon ved oppdagelse.
- **Presedens:** Prosjektets eksplisitte versjonskrav og ikke-hemmelige miljøvariabler gjelder for prosjektprosessen. Verdier som ikke er angitt, arves fra miljøet der det er hensiktsmessig og trygt; eksplisitte, dokumenterte kommandoalternativer kan overstyre for den aktuelle kjøringen. Maskinens konfigurasjonsfil endres ikke. Spesifiser og test lokale tilpasninger, hemmeligheter og konfliktregler før implementering. Dokumenter hvordan `.nvmrc`, `.node-version`, `package.json` og `mise.toml` oppdages, og hva som skjer ved motstrid; ingen eksisterende filer skal endres i det skjulte. Ikke tolk `package.json` `engines` som et eksakt versjonsvalg uten en dokumentert regel.
- **Versjonsoppløsning:** Skill ønsket versjon eller intervall fra faktisk versjon og installasjonskilde. Prosjektets valg påvirker prosessen og dens barn, ikke den globale maskinkonfigurasjonen. Vis tydelige feil for manglende versjon, utilgjengelig backend og uklar konflikt.
- **Backend:** Evaluer fnm, mise og eventuelt flere adaptere mot et felles grensesnitt for oppdagelse, versjonsoppløsning, installasjon og kjøring. Ikke installer eller bytt versjonsbehandler automatisk. Første leveranse må støtte minst én eksplisitt valgt og testet Node-backend; flere språk og backender følger etter dokumentert behov.
- **Oppgaver:** List oppgaver og vis hvilken kommando, arbeidsmappe og hvilke verktøy som vil brukes før kjøring. Videresend argumenter, signaler og avslutningskode korrekt. Definer eksplisitt om en oppgave kjøres via shell, og håndter sitering og plattformforskjeller.
- **Miljø:** Tillat eksplisitte, ikke-hemmelige verdier og definer arv og overstyring. `dk run` skal ikke endre forelder-shellets miljø. Hemmelige verdier skal ikke skrives til delte filer, logger eller standardutskrift; integrasjon med en ekstern hemmelighetsløsning vurderes separat.
- **Tillit:** Å lese eller liste en ukjent prosjektfil skal ikke kjøre kode, evaluere uttrykk eller installere programvare. Før første kjøring av en oppgave eller annet kodekjørende innhold skal brukeren se hva som godkjennes. Tillit knyttes til en konkret konfigurasjon og relevant innhold; vesentlige endringer krever ny vurdering. Ikke-interaktiv kjøring uten forhåndsgitt tillit skal feile tydelig.
- **Installasjon:** Hvis ønsket versjon mangler, vis plan og kilde. Ingen skjult installasjon ved katalogbytte eller ren inspeksjon. Eventuell installasjon krever et uttrykkelig valg eller en dokumentert ikke-interaktiv policy.
- **Filer og reproduksjon:** Vurder låsefil for eksakte versjoner og kilder. Beskriv hva som faktisk er låst og hva som fortsatt styres av eksterne pakkebehandlere; ikke lov identiske bygg uten grunnlag.
- **Synlighet:** Tilby en inspeksjon som viser funnet konfigurasjon, aktiv versjon, valgt backend, kilder og eventuelle konflikter. Eksakt kommando og flagg fastsettes etter brukerutprøving.

### Ferdig når

To uavhengige repoer med henholdsvis Node 22 og Node 24 kan klargjøres med uttrykkelig valg og kjøres etter hverandre med `dk run`, også fra en undermappe. Et testskript i hvert repo dokumenterer faktisk Node-versjon, arbeidsmappe og egne ikke-hemmelige variabler; disse samsvarer med repoets krav, og endres ikke av det andre repoet. Den globale Node-standarden forblir uendret. Tester dekker nytt oppsett, gjenbruk, manglende versjon, flere Node-installasjoner på `PATH`, konflikter, ukjent prosjekt, endret tillit, ikke-interaktiv kjøring, avbrudd, argumenter og exit-koder. Inspeksjon installerer ingenting og kjører ikke prosjektkode. Hemmeligheter vises ikke i Datakraftens utdata eller logger; oppgaver brukeren kjører kan selv håndtere og vise egne miljøverdier.

## 8. Produktområde: Sikkerhet som løpende støtte

### Hvorfor

Utviklerverktøy kjører ofte med tilgang til kildekode, miljøvariabler og legitimasjon. Ondsinnede pakker og kompromitterte avhengigheter kan derfor ramme både arbeidsstasjonen og systemene utvikleren har tilgang til. [OpenSSF beskriver risikoen ved installasjonsskript og arbeidsstasjoner](https://openssf.org/blog/2026/07/28/what-is-a-dependency-firewall/); [CISA omtaler sikring av utviklingsmiljøer og kontoer](https://www.cisa.gov/sites/default/files/2023-12/ESF_SECURING_THE_SOFTWARE_SUPPLY_CHAIN_DEVELOPERS.pdf).

Datakraften skal hjelpe med å oppdage og redusere konkrete risikoer innenfor sitt dokumenterte dekningsområde. Det skal aldri gi inntrykk av at en maskin er «sikker» fordi en kontroll ikke fant noe.

### Leveranser i prioritert rekkefølge

**1. Synlighet og eierskap**

- Vis verktøyets registrerte navn, versjon, plassering, installasjonskilde og eier når informasjonen finnes. Ukjent kilde eller eierskap vises som ukjent.
- Vis hvilken binær som faktisk kjører, og mulige `PATH`-konflikter. Ikke kjør vilkårlige ukjente binærer bare for å skanne dem uten å dokumentere og vurdere konsekvensen.
- Oppdag manglende eller avsluttet støtte for utvalgte runtimes der Datakraften har pålitelige versjons- og livssyklusdata. Oppgi kilde og dato; skill mellom ukjent, støttet og utløpt støtte.

**2. Gjennomtenkte installasjoner og oppdateringer**

- Vis kilde, pakke, valgt versjon, nødvendige privilegier og hva som faktisk skal endres før installasjon eller oppdatering.
- Verifiser integritet og opphav når relevant kilde tilbyr verifiserbar informasjon. Skill tydelig mellom sjekksum, signatur, attestering og ikke verifisert kilde. Kontroller at tillitsnøkler og forventede signaturer selv har et dokumentert tillitsgrunnlag.
- Dokumenter risikoen ved installasjonsskript og transitive avhengigheter. En tilgjengelig, gyldig signatur bekrefter ikke at innholdet er skadefritt.
- Planlegg sikrere håndtering av npm-verktøy og behovet for sudo, uten å endre brukerens prefix eller versjonsbehandler i det skjulte. Brukeren skal få forklart om en handling kjører installasjonsskript med forhøyede rettigheter.
- Endringer i kilder, leverandør eller vesentlige rettigheter skal synliggjøres før de anvendes når de kan oppdages pålitelig.

**3. Konkrete sikkerhetsfunn**

- Evaluer et begrenset, dokumentert første dekningsområde for kjente sårbarheter og skadelige pakker, med pålitelige kilder som [OSV](https://osv.dev/) og leverandørenes sikkerhetsmeldinger. Kartlegg økosystem, pakkenavn og versjon nøyaktig før oppslag.
- Hvert funn skal vise berørt verktøy og versjon, kilde, tidspunkt, relevant identifikator, alvorlighetsgrad der den finnes, og en praktisk neste handling eller lenke til leverandøren.
- «Ingen kjente funn» gjelder bare kontrollerte kilder, versjoner og tidspunkt. Uidentifisert pakke, manglende nettverk, utdatert database, ufullstendig økosystemstøtte og oppslagsfeil skal vises som **ukjent / ikke kontrollert**, ikke grønt.
- Skil mellom varsling, anbefaling og blokkering. Automatisk blokkering krever tydelige, testede kriterier og dokumentert håndtering av feilaktige treff og unntak. Ingen automatisk tvungen oppdatering i bakgrunnen.
- Prosjektavhengigheter, editorutvidelser og AI-tilleggsprogrammer omfattes bare når de uttrykkelig støttes og faktisk kan identifiseres. Ufullstendige resultater må merkes.

**4. Sikker bruk over tid**

- `dk doctor` og planlagte sikkerhetsvisninger skal prioritere det brukeren kan handle på, forklare konsekvens og skille kritiske funn, advarsler, hoppet over og ukjent.
- Oppdateringer skal kunne forhåndsvises og kjøres selektivt der backend støtter det. Rapporter hva som ble forsøkt, hva som lyktes og hva som fortsatt er ukjent; unngå falske påstander om atomaritet eller reversering.
- Undersøk valgfri varsling ved nye relevante funn etter at personvern, kilder og driftsmodell er definert. Første 3.0-krav er en eksplisitt kontroll, ikke en bakgrunnstjeneste.
- Ikke les, skriv ut eller send nøkler, passord, tokens eller hemmelige miljøvariabler. Ikke samle inventar eller prosjektdata på en server uten tydelig samtykke og spesifisert formål.

### Sikre Datakraften selv

- Vedlikehold en trusselmodell for prosjektfiler, fjernstyrt teamkonfigurasjon, CLI-kommandoer, oppdateringer, bygg og utgivelser. Test relevante misbrukstilfeller.
- Styrk utgivelseskjeden med gjennomgåtte CI-rettigheter, avhengighetsoppdateringer og en dokumentert strategi for signering/attestering. Definer hvordan nøkkel- og tillitsrot verifiseres, distribueres og roteres før signering påstås som levert.
- Oppdater `SECURITY.md` slik at sikkerhetsløfter, rapporteringsvei og støttede versjoner stemmer med faktisk praksis. Vedlikehold ansvar og prosess for sikkerhetsmeldinger.
- Hent aldri fjernkonfigurasjon eller oppdateringer som kjøres automatisk uten en dokumentert tillitsmodell, gjennomgang av endringer og definerte feiltilstander.

### Ferdig når

Det første dekningsområdet er eksplisitt dokumentert og testet fra deteksjon til funn og tiltak. Uten nettverk eller ved ukjent kilde presenteres ikke resultatet som en bestått sikkerhetskontroll. Hemmeligheter lekker ikke til testet utdata eller logger. Prosjektets egen utgivelsessikkerhet og varsling om sårbarheter har en dokumentert, praktiserbar prosess.

## 9. Produktområde: Initialisering, deling og tillit

### Beslutning: Felles maskinoppsett importeres én gang

`dk init` skal kunne ta en parameter med adressen til en YAML-fil i et Git-repository. Filen er en **valgfri startmal for maskinen**, ikke en løpende fjernstyrt konfigurasjon. Når brukeren initialiserer, leser `dk` filen én gang, fletter innholdet med Datakraftens standarder og lagrer resultatet som en fullstendig, lokal og redigerbar maskinkonfigurasjon. Eksterne innstillinger har forrang; standarder fyller inn felter som mangler. Etterpå er den lokale filen kilden til sannhet: brukeren kan endre den når som helst, og senere `dk apply`, `dk update` og andre ordinære kommandoer skal ikke hente malen på nytt.

```bash
# Foreslått parameter og URL-format; eksakt CLI-syntaks er ikke implementert
dk init --from https://raw.githubusercontent.com/eksempel/oppsett/main/datakraften.yaml
dk apply
```

`dk init` oppretter konfigurasjonen; det installerer ikke verktøy som en uannonsert bivirkning. Brukeren kan gjennomgå og redigere resultatet før `dk apply`. Om en fremtidig import støtter flere typer Git-leverandøradresser, private repoer eller en bestemt ref, skal dette dokumenteres og testes per støttet kilde. En vanlig nettleserside med HTML må ikke stille aksepteres som YAML; en rå filadresse er første tydelige mål. Ved tilgangs- eller henteproblemer skal brukeren få en handlingsrettet feil uten at legitimasjon, URL-tokens eller hemmeligheter logges.

### Flettekontrakt for lokal maskinkonfigurasjon

1. Generer samme plattformtilpassede standard som `dk init` uten ekstern fil ville brukt.
2. Hent og parse den eksterne YAML-filen som en partiell maskinkonfigurasjon. Valider tillatte felter, typer, format og versjon før noe skrives eller kjøres.
3. Flett objekter rekursivt: felter som finnes i den eksterne filen, overstyrer standardverdiene på samme nøkkel; manglende felter beholder standardverdiene. Eksplisitt `false`, `0` og tom streng er verdier, ikke manglende felter. Lister som er angitt, erstatter standardlisten; en eksplisitt tom liste blir tom. Et tomt objekt overstyrer ikke hele et eksisterende standardobjekt. `null` avvises med forklaring inntil en egen semantikk er definert, og ugyldige typer eller ukjente nøkler avvises med en presis feil.
4. Valider den fullstendige flettede konfigurasjonen, vis kilden og et forståelig resultat eller en oversikt over endringene, og lagre først deretter atomisk i `~/.config/datakraften/config.yaml`. Ikke lagre hemmeligheter i kildemetadata eller vise dem i utdata.
5. Den lagrede konfigurasjonen skal representere et lokalt redigerbart oppsett, med eksisterende `source: custom`-semantikk der den passer, og aldri en aktiv `source: team`-peker som fører til automatisk ny henting. Eventuell informasjon om hvor malen kom fra er kun metadata, uten automatisk synkronisering.

Eksempel: Dersom standarden aktiverer Node og Python, mens den eksterne filen bare setter `runtimes.node.enabled: false`, skal den lokale filen beholde standardene for Python og øvrige Node-felter, men uttrykkelig slå av Node. Endrer utvikleren senere Python lokalt, skal den endringen fremdeles gjelde neste gang `dk apply` kjøres, selv om den eksterne filen er endret eller utilgjengelig.

Eksisterende lokal konfigurasjon må aldri overskrives lydløst: `dk init` skal oppdage den, vise hva som vil skje og kreve et eksplisitt valg før eventuell erstatning. En mislykket henting, parsing, validering eller skriving skal gi en feil og bevare eksisterende konfigurasjon og tilstand. Ved repetert import skal `dk` ikke automatisk overskrive lokale tilpasninger eller stille slå sammen med en eksisterende fil; gjenimport krever et tydelig valg og en forhåndsvisning.

### Overgang fra dagens `source: team`

I dagens implementasjon lagrer `dk init --team` en URL-peker, og `dk apply` henter konfigurasjonen på nytt. Dette er nåværende atferd, ikke det nye målet. Eksisterende brukere skal ikke få endret oppsettet sitt lydløst: behold lesing og dagens virkemåte til en dokumentert, testet overgang er tilgjengelig. Tilby en eksplisitt migrering som henter og validerer valgt eksisterende kilde, viser flettingen med standardene, bevarer en sikkerhetskopi og skriver en lokal redigerbar konfigurasjon først etter brukerens valg. Etter vellykket migrering skal ordinære kommandoer ikke lenger hente fjernfilen. Ved utilgjengelig kilde eller mislykket migrering beholdes den opprinnelige filen og brukeren får en forklaring. Eksakt kommandonavn og eventuell utfasingsperiode fastsettes før lansering.

### Repoets prosjektfil

- `datakraften.yaml` i repoet er prosjektets delte krav til versjoner, prosjektmiljø og oppgaver. Den er ikke en fjernmal for arbeidsstasjonen og skal ikke lagres over maskinkonfigurasjonen.
- Repoets konfigurasjon følger koden og kan gjennomgås i pull requests; to repoer fra samme team kan angi ulike versjoner og variabler. Krav til tillit, prosjektoppdagelse og prosessavgrensning er beskrevet i kapittel 7.
- Prosjektspesifikke verdier overstyrer relevante maskin- og miljøstandarder **kun for prosjektprosessen** etter de dokumenterte presedensreglene. `dk init` skal alltid initialisere maskinkonfigurasjonen, selv om en prosjektfil finnes i gjeldende mappe.
- En prosjektfil er ikke en sikkerhetssandkasse. Før prosjektoppgaver eller installasjoner kjøres, må brukeren forstå og godkjenne kodekjørende innhold og kilder. Ingen hemmeligheter skal legges i delte filer.

### Akseptansekriterier for init og prosjekt

- [ ] `dk init` uten adresse gir ordinære plattformstandarder; med en gyldig partiell ekstern fil gir den en komplett lokal fil med korrekte overstyringer og utfylte standarder.
- [ ] Tester dekker rekursiv fletting, manglende felter, eksplisitt `false`, `0`, tom streng, tomme objekter og lister, ugyldige typer, `null` og ukjente felter.
- [ ] En lokal redigering overlever gjentatt `dk apply`; verken endret eller utilgjengelig ekstern fil påvirker den importerte konfigurasjonen uten en ny, eksplisitt import.
- [ ] Utilgjengelig kilde, HTML, ugyldig YAML og skrivefeil endrer ikke eksisterende konfigurasjon. Reinitialisering og migrering bevarer brukerens eksisterende data ved avslag eller feil.
- [ ] Et eksisterende `source: team`-oppsett fortsetter å fungere til brukeren velger migrering; migrert oppsett bruker den lokale filen og henter ikke på nytt.
- [ ] `datakraften.yaml` oppdages i gjeldende mappe og fra undermapper innenfor repoets grense; to repoer kan kjøre ulike Node-versjoner, variabler og oppgaver uten å overstyre hverandre eller maskinens globale standard.
- [ ] Inspeksjon, `dk init` og `dk apply` kjører aldri en prosjektoppgave bare fordi en prosjektfil finnes; manglende fil betyr ordinær maskinatferd uten feil.
- [ ] Hente-, tillits- og feilmeldinger eksponerer ikke passord, tokens eller hemmelige miljøvariabler; dokumentasjonen forklarer nøyaktig hva som er importert og hva som ikke synkroniseres.

## 10. Plattform og kompatibilitet

| Miljø | Mål før 3.0 | Hvordan dette bevises |
| --- | --- | --- |
| Fedora Linux og Ubuntu/Debian Linux | Testede primære miljøer | Nytt oppsett og eksisterende oppsett, Bash og valgt versjonsbackend, `apply`, `doctor`, `update` og prosjektkjøring |
| WSL på Ubuntu/Debian | Testet primærflyt | Reell WSL-kjøring, integrasjoner og tydelig skille mellom Linux- og Windows-side |
| WSL på Fedora | Støttes når testet | Dokumentert ende-til-ende-test; ellers tydelig markert som ikke verifisert |
| macOS på amd64 og arm64 | Eksakt funksjonsnivå avgjøres før 3.0 | CI-bygg på begge arkitekturer og reelle tester på tilgjengelige macOS-miljøer; krysskompilering alene er ikke funksjonsbevis |
| Native Windows | Ikke et 3.0-krav | WSL kan brukes der støttet; ingen løfter om native funksjoner |

For hver funksjon dokumenteres status som testet, delvis, ikke støttet eller ikke verifisert. Plan for oppgradering fra eksisterende `dk`-konfigurasjoner skal inkludere forhåndsvisning, sikkerhetskopi ved nødvendige filendringer og tydelig mulighet til å beholde gjeldende oppsett. Eksisterende kommandoer og `source`-verdier skal bevares eller få en dokumentert migrering før en eventuell endring.

## 11. Leveransefaser og prioriteringer

Fasene er avhengighetsstyrte, ikke kalenderløfter. Versjonsnumre før 3.0 og eksakte API-er bestemmes når omfang og kompatibilitet er avklart. Hver fase kan leveres inkrementelt i mindre utgivelser.

### Fase 0 – Gjør retningen etterprøvbar

**Prioritet: først.**

- Bekreft målgrupper og gjennomfør samtaler med utviklere om oppdagelse, prosjektbytte, vedlikehold og sikkerhet.
- Lag en funksjons- og plattformmatrise basert på kode og tester. Rett de viktigste avvikene mellom dokumentasjon og produkt.
- Definer sikkerhetsmodell, eierskap til installasjoner, målemetoder for katalogen og avgrenset dekningsområde.
- Beslutning: hva er minste reelle 3.0-opplevelse, og hvilke eksisterende installasjoner skal støttes uten overtakelse?

**Utgangskriterium:** Dokumentert baseline og prioriterte problemer, med kjent testdekning og avklarte åpne beslutninger.

### Fase 1 – Fundament og pålitelig arbeidsstasjon

**Prioritet: må på plass før prosjektkjøring og sikkerhetsløfter bygges oppå.**

- Pålitelig konfigurasjonsvalidering, feilmeldinger, tilstandsmodell og konsekvent oversikt over aktive og administrerte verktøy.
- Implementer `dk init` med engangsimport fra ekstern Git-YAML, standardutfylling og lokal eierskap etter flettekontrakten i kapittel 9. Test alle feiltilstander før lokal fil erstattes.
- Etabler og test eksplisitt migrering fra eksisterende `source: team`, med forhåndsvisning og sikkerhetskopi; bevar gammel atferd til brukeren velger migrering.
- Sikre, idempotente endringer og reelle forhåndsvisninger. Test eksisterende shelloppsett og PATH-konflikter.
- Korriger bredde og ordlyd for `dk update` slik at brukeren vet når andre pakker berøres.
- Etabler trusselmodell, prosjektets sikkerhetsprosess og testbare tillitsgrenser.

**Utgangskriterium:** En bruker kan se hva `dk` finner og planlegger, gjenta oppsettet og forstå feil uten misvisende suksessmeldinger.

### Fase 2 – Prosjektmiljø: første sammenhengende flyt

**Prioritet: første produktutvidelse.**

- Oppdag `datakraften.yaml` i gjeldende mappe og repoets undermapper etter reglene i kapittel 7; innfør skjemaversjon og avklart samspill med eksisterende prosjektfiler. Oppdagelse skal ikke utløse kommandoer eller installasjoner.
- En sammenhengende flyt for prosjektinspeksjon, uttrykkelig oppsett og `dk run`, med én testet Node-adapter, prosjektlokalt prosessmiljø, ikke-hemmelige miljøvariabler og navngitte oppgaver.
- Oversikt over planlagt kjøring, tillit for nytt eller endret kodekjørende innhold og eksplisitt håndtering av manglende versjoner.
- Integrasjonstester for to uavhengige repoer med Node 22 og 24, egne miljøvariabler og oppgaver, samt feil, avbrudd, ukjente prosjektfiler og eksisterende versjonsbehandler.

**Utgangskriterium:** Akseptansekriteriet i kapittel 7 er oppfylt på prioriterte miljøer. Beslutning om ytterligere backender og språk tas ut fra faktisk bruk.

### Fase 3 – Katalog og oppdagelse

**Prioritet: parallelt med eller etter fase 2 når datagrunnlaget er klart.**

- Publiser en første kuratert katalog med minst to kategorier og presis status for støttede installasjoner.
- Søk, sammenligninger og dokumenterte kilder. Utforsk første brukertest av veien fra oppdagelse til installasjon.
- Innfør eventuelle popularitetslister først når datakvalitet, oppdateringsfrekvens og metode kan vises for brukeren.

**Utgangskriterium:** Brukertest viser at utviklere kan finne, forstå og velge et verktøy uten å forveksle popularitet, anbefaling og sikkerhet.

### Fase 4 – Sikkert vedlikehold

**Prioritet: på fundamentet fra fase 1, med begrenset første dekningsområde.**

- Inventar og oppdatert kildeinformasjon for utvalgte verktøy.
- Én gjennomarbeidet flyt for oppslag mot en valgt sikkerhetskilde, med tydelige grenser, ukjent-tilstander og handlingsrettet resultat.
- Forhåndsvis og utfør utvalgte oppdateringer med korrekte privilegier og tydelig omfang.
- Dokumenter og gjennomfør tiltak for prosjektets egen utgivelseskjede og sårbarhetsrapportering.

**Utgangskriterium:** Akseptansekriteriet i kapittel 8 er oppfylt, inkludert feil i nettverk og datakilder. Ingen markedsføring om fullstendig beskyttelse.

### Fase 5 – Integrasjon, migrering og 3.0

- Sammenhengende brukerreise fra katalog til maskin, prosjekt, diagnostikk og vedlikehold.
- Dokumentert migrering fra støttede tidligere versjoner, revisjon av CLI-/JSON-kontrakter og gjennomgang av eksempler og nettside.
- Kompatibilitets- og sikkerhetstesting, brukertester på reelle miljøer og utbedring av kritiske mangler.
- Publiser en presis støttematrise, kjente begrensninger, endringslogg, utgivelsesnotater og rutine for feilrapportering.

**Utgangskriterium:** Alle 3.0-kriteriene nedenfor er oppfylt. Utsett funksjoner eller versjon fremfor å gi feil trygghet.

## 12. Krav for å kalle utgivelsen 3.0

### Må være oppfylt

- [ ] Visjonen og løftene på nettside og i dokumentasjon stemmer med det som faktisk er tilgjengelig.
- [ ] En dokumentert og testet brukerreise dekker oppdagelse, oppsett, prosjektkjøring, diagnostikk og vedlikehold. Katalogen trenger ikke dekke alle verktøy.
- [ ] Prosjektinspeksjon, oppsett og kjøring oppfyller kontrakten og akseptansekriteriene i kapittel 7, med minst én testet Node-backend og to repoer med ulike versjoner.
- [ ] `dk init` oppfyller flette-, eierskaps-, feil- og migreringskravene i kapittel 9: ekstern YAML er en engangsimport, standarder fyller manglende felter, og resultatet er en lokal fil som brukeren kan endre.
- [ ] Sikkerhetsfunksjonene oppfyller et begrenset, offentlig dokumentert dekningsområde med eksplisitt håndtering av ukjent, feil og utdatert informasjon.
- [ ] Eksisterende brukere har en testet migreringsvei, og maskinens eksisterende verktøy og ikke-administrerte filer respekteres.
- [ ] Testmatrisen dekker prioriterte distroer, WSL og aktuelle shell, med dokumenterte begrensninger for andre plattformer.
- [ ] Go-tester, relevante integrasjons- og PTY-tester, bygg, Linux- og macOS-CI samt relevante nettstedskontroller er grønne for utgivelsen.
- [ ] Alle fire utgivelsesbinærer, SHA256-filer og installasjonsskript er publisert og kontrollert. Ytterligere signering/attestering omtales kun hvis den faktisk er innført og verifisert.
- [ ] Dokumentert prosess for private sikkerhetsmeldinger og respons er på plass og bemannet.
- [ ] Ingen kjente uavklarte feil som kan føre til utilsiktet kodekjøring, tap av brukerdata, lekkasje av hemmeligheter eller misvisende sikkerhetsstatus i 3.0-flytene.

### Bør leveres hvis kapasitet og data tillater det

- Flere prosjektbackender og språk enn første Node-flyt.
- Flere katalogkategorier og åpne, etterprøvbare popularitetsmålinger.
- Videre sikkerhetsdekning, automatiserte varsler og teamfunksjoner med dokumentert personvern.

### Utenfor 3.0 med mindre planen endres

- Fullstendig kartlegging eller sikring av alle installerte apper, editorutvidelser og prosjektavhengigheter.
- Automatisk bakgrunnsoppdatering som standard, universell rollback eller garanti for identiske maskiner.
- Egen pakkebehandler, obligatorisk skyinnlogging, omfattende administrasjonsplattform eller native Windows-støtte.

## 13. Måling og læring

Målene under er forslag til evaluering, ikke påståtte resultater. Sett tallfestede mål etter at en baseline er etablert.

| Hva vi vil lære | Måling eller test | Begrensning |
| --- | --- | --- |
| Kan brukere finne riktig verktøy? | Brukertest: finn og sammenlign to alternativer og forklar valget | Et begrenset testutvalg representerer ikke alle utviklere |
| Er oppsettet forståelig? | Andel testpersoner som korrekt kan beskrive hva planen vil endre | Gjennomføres på representative nye og eksisterende miljøer |
| Fungerer prosjektbytte? | Andel testprosjekter som kjører med forventet versjon og miljø | Registrer backend, shell og plattform; ikke skjul feilende kombinasjoner |
| Gir sikkerhetsfunn verdi? | Test om brukeren forstår kilde, usikkerhet og neste handling | Antall varsler eller fravær av varsler er ikke et mål på sikkerhet |
| Er vedlikehold forutsigbart? | Test forhåndsvisning mot faktiske endringer og delvis feil | Underliggende pakkebehandlere kan ha ulik kontroll og reversering |
| Er Datakraften foretrukket? | Frivillige intervjuer, gjentatt bruk og dokumenterte tilbakemeldinger | Ingen skjult telemetri; aktivitet er ikke alene bevis for preferanse |

## 14. Åpne beslutninger før implementering

1. **Prosjektformat:** Filnavnet `datakraften.yaml` er besluttet. Hvordan versjoneres skjemaet, hvordan løses konflikter med eksisterende prosjektformater, og hvordan håndteres nestede repoer og lokale tilpasninger i detalj?
2. **Versjonsbackend:** Skal første Node-flyt bygge på eksisterende fnm-integrasjon, støtte eksisterende mise, eller tilby begge? Beslutningen krever en test av konflikter og brukeropplevelse.
3. **Prosjekttillit:** Hvilket konkret innhold inngår i en godkjenning, hvor lagres den, og hvordan trekkes den tilbake? Hvilke handlinger tillates i CI?
4. **Miljøvariabler og hemmeligheter:** Hva er standard for arv, overstyring, skjuling og `.env`-filer? Hvilke eksterne hemmelighetsløsninger er relevante uten å gjøre én obligatorisk?
5. **Sikkerhetsdekning:** Hvilke konkrete verktøy/økosystemer og datakilder skal støttes først, og hvilke forhold gjør en kontroll ukjent?
6. **Katalog:** Hvilke to kategorier gir mest verdi først? Hvem kvalitetssikrer dataene, og hvordan håndteres rettelser og utdatert informasjon?
7. **Init og eldre teamoppsett:** Engangsimport og lokalt eierskap er besluttet. Hva skal init-parameteren og migreringskommandoen hete, hvilke Git-leverandør-URL-er og private kilder støttes, og hva er en rimelig utfasingsperiode for gamle fjernpekere?
8. **Plattform:** Hvilke macOS- og WSL-miljøer kan faktisk testes før 3.0, og hvem har ansvar for testene?
9. **Vedlikehold:** Hvilke eksisterende `dk update`-operasjoner kan avgrenses trygt, og hvor må bredere oppdatering forklares eller kreve særskilt valg?
10. **Kapasitet og styring:** Hvem eier hvert område, hvem kan godkjenne sikkerhets- og produktbeslutninger, og hva må tas ut dersom kapasiteten ikke strekker til?

Disse spørsmålene skal avklares gjennom små tekniske undersøkelser og brukertester før løsninger låses. Veikartet kan justeres når ny kunnskap tilsier det.
