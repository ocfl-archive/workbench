---
sessionId: session-260917-142226-1mb0
---

# Requirements

### Overview & Goals
Ziel ist es, den Workflow-Nutzen und die Effizienz bei der Stapelverarbeitung zu maximieren: Nach dem Schließen des Prozess-Modalfensters (`execModal`) sollen die Auswahlen in der Batch-Liste und der Job-Liste exakt auf den zuvor selektierten Elementen erhalten bleiben, während die Daten im Hintergrund aktualisiert und die Detailansicht sowie Aktions-Buttons nahtlos an den neuen Zustand angepasst werden.

### Scope
- **In Scope:**
  - Erfassung und Speicherung des aktuellen Auswahlzustands (`batchList` Index/Name sowie `jobList` Index/Signatur) vor der Aktualisierung.
  - Gezieltes Wiederherstellen der Selektion in `jobList` nach dem Neu-Einlesen der Jobs eines Batches.
  - Automatische Synchronisation der Detailansicht (`showDetail`) und der dynamischen Aktions-Buttons (`updateDetailAndActions`) für den beibehaltenen Job.
  - Wiederherstellung des aktiven UI-Fokus (z. B. auf `jobList` oder den aktualisierten Aktions-Button) nach dem Schließen des Modals (`pages.RemovePage("execModal")`).
- **Out of Scope:**
  - Persistierung der Selektion über Anwendungsneustarts hinweg auf der Festplatte.
  - Änderungen am Prozessaufruf oder der Logausgabe selbst.

### User Stories
- **Als Benutzer** möchte ich nach Ausführung einer Aktion (z. B. OCFL- oder Report-Erstellung) direkt an derselben Stelle in der Job-Liste weiterarbeiten, **damit** unnötige Klicks, Navigationszeit und kognitive Reibung minimiert werden.

### Functional Requirements
- Beim Schließen des Modalfensters (`execModal`) wird der aktuell gewählte Batch beibehalten.
- Der zuvor ausgewählte Job in der Job-Liste wird anhand seines Index bzw. seiner Signatur wieder als aktives Element markiert (`jobList.SetCurrentItem`).
- Die Detailansicht und die Aktions-Buttons spiegeln unmittelbar den aktualisierten Dateistatus des erhaltenen Jobs wider (z. B. Wechsel von "OCFL erstellen" zu "Report erstellen" oder "Vollständig").
- Der Tastaturfokus wird sauber auf das Hauptfenster (z. B. `jobList` oder die Detail-/Aktionsansicht) zurückgesetzt, sodass sofort nahtlos weiter navigiert werden kann.

### Non-Functional Requirements
- **Workflow-Optimierung:** Minimierung von Benutzerinteraktionen zur Wiederherstellung des Kontexts.
- **Robustheit:** Falls sich nach einem Refresh die Anzahl der Jobs ändert oder der vorherige Job nicht mehr existiert, greift ein sicheres Fallback (nächstliegender Index bzw. Index 0).

# Technical Design

### Current Implementation
- Beim Schließen des Modals wird `onFinish()` aufgerufen, welches `refreshCurrentBatch()` ausführt.
- `refreshCurrentBatch()` ruft `loadJobsForBatch(batchName)` auf:
  ```go
  loadJobsForBatch := func(batchName string) {
      jobList.Clear()
      detailView.Clear()
      buttonFlex.Clear()
      actionButton = nil
      currentJobs = nil
      ...
      currentJobs = jobs
      for _, job := range jobs {
          jobList.AddItem(job.signature, "", 0, nil)
      }
      if len(currentJobs) > 0 {
          updateDetailAndActions(currentJobs[0]) // Setzt immer auf Index 0 zurück!
      }
  }
  ```
- Dadurch verliert `jobList` die ursprüngliche Position und setzt stets das erste Element aktiv.

### Key Decisions
1. **Zustandserfassung vor Refresh:**
   - Vor dem Leeren von `jobList` wird der aktuelle Index `savedJobIndex := jobList.GetCurrentItem()` bzw. die Signatur des selektierten Jobs festgehalten.
2. **Index-Restaurierung mit Validierungsprüfung:**
   - Nach dem erneuten Befüllen von `jobList` wird geprüft, ob `savedJobIndex` innerhalb des gültigen Wertebereichs (`0 <= savedJobIndex < len(currentJobs)`) liegt.
   - Falls gültig, wird `jobList.SetCurrentItem(savedJobIndex)` aufgerufen und `updateDetailAndActions(currentJobs[savedJobIndex])` ausgeführt.
3. **Fokus-Rückgabe bei Modal-Schließung:**
   - Nach `pages.RemovePage("execModal")` wird der Fokus explizit auf `jobList` (oder die Detail-/Buttonansicht) gesetzt, um verwaiste Fokus-Zustände zu vermeiden.

### Proposed Changes

#### 1. Anpassung von `loadJobsForBatch` / `refreshCurrentBatch`
```go
loadJobsForBatch := func(batchName string, preserveIndex ...int) {
    targetIndex := 0
    if len(preserveIndex) > 0 && preserveIndex[0] >= 0 {
        targetIndex = preserveIndex[0]
    } else if currentIdx := jobList.GetCurrentItem(); currentIdx >= 0 {
        targetIndex = currentIdx
    }

    jobList.Clear()
    detailView.Clear()
    buttonFlex.Clear()
    actionButton = nil
    currentJobs = nil

    if batchName == "" || *inputFolder == "" {
        return
    }
    batchPath := path.Join(*inputFolder, batchName)
    jobs, err := getSignatures(batchPath, *ocflFolder, *reportFolder, logger)
    if err != nil {
        logger.Error().Err(err).Msgf("Failed to get jobs for batch %s", batchName)
        return
    }
    currentJobs = jobs
    for _, job := range jobs {
        jobList.AddItem(job.signature, "", 0, nil)
    }

    if len(currentJobs) > 0 {
        if targetIndex >= len(currentJobs) {
            targetIndex = len(currentJobs) - 1
        }
        jobList.SetCurrentItem(targetIndex)
        updateDetailAndActions(currentJobs[targetIndex])
    }
}

refreshCurrentBatch = func() {
    if batchIndex := batchList.GetCurrentItem(); batchIndex >= 0 {
        batchName, _ := batchList.GetItemText(batchIndex)
        jobIndex := jobList.GetCurrentItem()
        loadJobsForBatch(batchName, jobIndex)
    }
}
```

#### 2. Modal Close Handler & Fokus-Rückstellung
```go
modalText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
    pages.RemovePage("execModal")
    if onFinish != nil {
        onFinish()
    }
    app.SetFocus(jobList)
    return nil
})
```

### File Structure
- `cmd/create_container/main.go` (modifiziert)

# Testing

### Validation Approach
- Kompilierung und Code-Inspektion zur Sicherstellung der Typsicherheit.
- Verifikation des Erhalts des Auswahlindexes beim Triggern des Refresh-Zyklus.

### Key Scenarios
1. **Selektionserhalt bei Refresh:**
   - Einen Batch auswählen und einen Job an mittlerer/hinterer Position (z. B. Index 2 oder 3) selektieren.
   - Aktion ausführen und das Modal schließen.
   - Sicherstellen, dass genau dieser Job in `jobList` selektiert bleibt und seine Detailansicht aktualisiert ist.
2. **Fokus-Wiederherstellung:**
   - Prüfen, dass nach Schließen des Modalfensters sofort Tastatur-Navigation (Pfeiltasten, Tab) ohne vorherigen Mausklick möglich ist.

### Edge Cases
- Leere Batches oder Batches mit 0 Jobs.
- Ein Job entfällt nach der Erstellung unerwartet aus der Liste: Index wird auf den maximal verfügbaren Index begrenzt, ohne Panic.

# Delivery Steps

### ✓ Step 1: Selektionserhalt und Fokuswiederherstellung implementieren
Die Auswahl in der Batch- und Job-Liste sowie der UI-Fokus bleiben nach dem Schließen des Modalfensters und Daten-Refresh erhalten.

- `loadJobsForBatch` so anpassen, dass optional ein Zielindex übergeben oder der aktuelle `jobList`-Index beibehalten wird.
- `refreshCurrentBatch` erweitern, sodass der aktuelle Index an `loadJobsForBatch` übergeben wird.
- Im Schließen-Handler von `execModal` den Fokus nach `pages.RemovePage` explizit auf `jobList` setzen.

### ✓ Step 2: Build und Verhalten verifizieren
Kompilierung und Code-Inspektion zur Sicherstellung der fehlerfreien Ausführung.

- Mittels `go build ./cmd/create_container` die fehlerfreie Kompilierung sicherstellen.
- Code auf Robustheit bei leeren Batches oder Indexgrenzen prüfen.