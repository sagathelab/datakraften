# Datakraften 3.0 – veikart

**Status:** Produktretning og plan, ikke en beskrivelse av funksjonalitet som allerede er levert.  
**Utgangspunkt:** `v0.12.23`, 11. november 2026.  
**Tidligere planer:** [2.0-veikartet](datakraften_2.0_roadmap.md) og [plattformplanen](datakraften-platform-plan.md) er historiske og inneholder både leverte og uleverte forslag. Dette dokumentet er retningen for arbeidet mot 3.0.

## 1. Visjon

> A better starting point for your development machine.  
> Datakraften brings your runtimes, shell, editors and AI tools together in a configuration you can understand and share.

Dette er fortsatt kjernen. Datakraften skal i tillegg være en **støttespiller gjennom hele utviklermiljøets livsløp**: fra å oppdage og forstå verktøy, via installasjon og oppsett, til prosjektarbeid, diagnostikk og trygt vedlikehold.

Målet er å bli en foretrukket løsning for utviklere fordi produktet er nyttig i hverdagen, forståelig, pålitelig og verdt tilliten. Det er en ambisjon som må vinnes gjennom faktiske brukerbehov og kvalitet – ikke noe en versjonsnummerering i seg selv oppnår.

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

Et prosjekt skal kunne beskrive verktøyversjoner, ikke-hemmelige miljøvariabler og navngitte oppgaver. En utvikler skal forstå og kjøre prosjektet uten å endre maskinens globale standardversjoner.

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
dk run dev
dk run test
```

### Kontrakt som må utformes og testes

- **Fil og søk:** Definer filnavn, skjemaversjon, prosjektrot, oppadgående søk og eksplisitt valg av fil. En prosjektfil og `~/.config/datakraften/config.yaml` har ulike roller; de skal ikke blandes eller overskrive hverandre.
- **Presedens:** Spesifiser og test hvordan eksplisitte flagg, prosjektfil, eventuelle overordnede prosjektfiler, eksisterende miljøvariabler og standardverdier virker sammen. Dokumenter hvordan `.nvmrc`, `.node-version`, `package.json` og `mise.toml` oppdages, og hva som skjer ved motstrid. Ikke tolk `package.json` `engines` som et eksakt versjonsvalg uten en dokumentert regel.
- **Versjonsoppløsning:** Skill ønsket versjon eller intervall fra faktisk versjon og installasjonskilde. Prosjektets valg påvirker prosessen og dens barn, ikke den globale maskinkonfigurasjonen. Vis tydelige feil for manglende versjon, utilgjengelig backend og uklar konflikt.
- **Backend:** Evaluer fnm, mise og eventuelt flere adaptere mot et felles grensesnitt for oppdagelse, versjonsoppløsning, installasjon og kjøring. Ikke installer eller bytt versjonsbehandler automatisk. Første leveranse må støtte minst én eksplisitt valgt og testet Node-backend; flere språk og backender følger etter dokumentert behov.
- **Oppgaver:** List oppgaver og vis hvilken kommando, arbeidsmappe og hvilke verktøy som vil brukes før kjøring. Videresend argumenter, signaler og avslutningskode korrekt. Definer eksplisitt om en oppgave kjøres via shell, og håndter sitering og plattformforskjeller.
- **Miljø:** Tillat eksplisitte, ikke-hemmelige verdier og definer arv og overstyring. `dk run` skal ikke endre forelder-shellets miljø. Hemmelige verdier skal ikke skrives til delte filer, logger eller standardutskrift; integrasjon med en ekstern hemmelighetsløsning vurderes separat.
- **Tillit:** Å lese eller liste en ukjent prosjektfil skal ikke kjøre kode, evaluere uttrykk eller installere programvare. Før første kjøring av en oppgave eller annet kodekjørende innhold skal brukeren se hva som godkjennes. Tillit knyttes til en konkret konfigurasjon og relevant innhold; vesentlige endringer krever ny vurdering. Ikke-interaktiv kjøring uten forhåndsgitt tillit skal feile tydelig.
- **Installasjon:** Hvis ønsket versjon mangler, vis plan og kilde. Ingen skjult installasjon ved katalogbytte eller ren inspeksjon. Eventuell installasjon krever et uttrykkelig valg eller en dokumentert ikke-interaktiv policy.
- **Filer og reproduksjon:** Vurder låsefil for eksakte versjoner og kilder. Beskriv hva som faktisk er låst og hva som fortsatt styres av eksterne pakkebehandlere; ikke lov identiske bygg uten grunnlag.
- **Synlighet:** Tilby en inspeksjon som viser funnet konfigurasjon, aktiv versjon, valgt backend, kilder og eventuelle konflikter. Eksakt kommando og flagg fastsettes etter brukerutprøving.

### Ferdig når

To testprosjekter med ulike Node-versjoner kan kjøres etter hverandre med `dk run`; hvert prosjekt får riktig versjon og miljø, testoppgavens exit-kode og signaler videreføres, og global Node-standard forblir uendret. Manglende backend, motstridende konfigurasjon, ukjent prosjekt og ikke-interaktiv kjøring testes eksplisitt. Ingen kode kjøres ved ren inspeksjon.

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

## 9. Produktområde: Deling, team og tillit

- Behold den eksisterende modellen med `source: default | custom | team`, og planlegg migrering før endringer i skjema eller semantikk.
- Delte konfigurasjoner skal kunne inspiseres og valideres før de påvirker maskinen. Vis opprinnelse, innhold/endringer og hvilke installasjoner eller oppgaver som følger.
- Definer hvordan teamkonfigurasjon festes til en bestemt versjon eller et innholdshash, hvordan oppdatering av tillit skjer, og hva som skjer når kilden er utilgjengelig eller innholdet er endret. HTTPS alene er ikke tilstrekkelig garanti for at innholdet er godkjent.
- Ingen hemmeligheter i delte filer. Beskriv integrasjonspunkter for eksisterende hemmelighetshåndtering uten å lage en egen hemmelighetstjeneste som 3.0-krav.
- Teamets anbefalinger skal kunne sameksistere med lokale valg gjennom dokumentert presedens og synlige avvik. Ikke hev at alle maskiner er identiske når kilder, versjoner eller plattformstøtte varierer.
- Før kommersielle teamfunksjoner eller sentral innsamling vurderes, dokumenter personvern, eierskap, nytte og eksplisitt samtykke.

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
- Sikre, idempotente endringer og reelle forhåndsvisninger. Test eksisterende shelloppsett og PATH-konflikter.
- Korriger bredde og ordlyd for `dk update` slik at brukeren vet når andre pakker berøres.
- Etabler trusselmodell, prosjektets sikkerhetsprosess og testbare tillitsgrenser.

**Utgangskriterium:** En bruker kan se hva `dk` finner og planlegger, gjenta oppsettet og forstå feil uten misvisende suksessmeldinger.

### Fase 2 – Prosjektmiljø: første sammenhengende flyt

**Prioritet: første produktutvidelse.**

- Prosjektfil med definert søk, presedens, skjemaversjon og en avklart håndtering av eksisterende prosjektfiler.
- Én testet Node-adapter, prosjektlokalt prosessmiljø, ikke-hemmelige miljøvariabler og navngitte oppgaver via `dk run`.
- Oversikt over planlagt kjøring, tillit for nytt eller endret kodekjørende innhold og eksplisitt håndtering av manglende versjoner.
- Integrasjonstester for to ulike Node-prosjekter, feil, avbrudd, ukjente prosjektfiler og eksisterende versjonsbehandler.

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
- [ ] Prosjektkjøring oppfyller kontrakten og akseptansekriteriene i kapittel 7, med minst én testet Node-backend.
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

1. **Prosjektformat og filnavn:** Kan `datakraften.yaml` brukes uten å forveksles med arbeidsstasjonskonfigurasjonen? Hva er kompatibilitetsreglene og presedensen?
2. **Versjonsbackend:** Skal første Node-flyt bygge på eksisterende fnm-integrasjon, støtte eksisterende mise, eller tilby begge? Beslutningen krever en test av konflikter og brukeropplevelse.
3. **Prosjekttillit:** Hvilket konkret innhold inngår i en godkjenning, hvor lagres den, og hvordan trekkes den tilbake? Hvilke handlinger tillates i CI?
4. **Miljøvariabler og hemmeligheter:** Hva er standard for arv, overstyring, skjuling og `.env`-filer? Hvilke eksterne hemmelighetsløsninger er relevante uten å gjøre én obligatorisk?
5. **Sikkerhetsdekning:** Hvilke konkrete verktøy/økosystemer og datakilder skal støttes først, og hvilke forhold gjør en kontroll ukjent?
6. **Katalog:** Hvilke to kategorier gir mest verdi først? Hvem kvalitetssikrer dataene, og hvordan håndteres rettelser og utdatert informasjon?
7. **Team:** Hvordan verifiseres og oppdateres fjernkonfigurasjon, og hvordan balanseres felles krav mot lokale valg?
8. **Plattform:** Hvilke macOS- og WSL-miljøer kan faktisk testes før 3.0, og hvem har ansvar for testene?
9. **Vedlikehold:** Hvilke eksisterende `dk update`-operasjoner kan avgrenses trygt, og hvor må bredere oppdatering forklares eller kreve særskilt valg?
10. **Kapasitet og styring:** Hvem eier hvert område, hvem kan godkjenne sikkerhets- og produktbeslutninger, og hva må tas ut dersom kapasiteten ikke strekker til?

Disse spørsmålene skal avklares gjennom små tekniske undersøkelser og brukertester før løsninger låses. Veikartet kan justeres når ny kunnskap tilsier det.
