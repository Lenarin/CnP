Dim appRef
Set appRef = CreateObject("Photoshop.Application")

Dim originalRulerUnits
originalRulerUnits = appRef.Preferences.RulerUnits
appRef.Preferences.RulerUnits = 2

Dim docRef
Dim artLayerRef
Dim textItemRef
Set docRef = appRef.Documents.Add(2, 4)

Set artLayerRef = docRef.ArtLayers.Add
artLayerRef.Kind = 2

Set textItemRef = artLayerRef.textItem
textItemRef.Contents = "Hello, World!"

appRef.Preferences.RulerUnits = originalRulerUnits