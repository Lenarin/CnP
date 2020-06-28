function pasteImage(filename, layerName, x, y, newWidth, newHeight) {

	function place(filename) {
		var desc = new ActionDescriptor();
		desc.putPath(stringIDToTypeID("null"), new File(filename));
		desc.putBoolean(stringIDToTypeID("unwrapLayers"), true);
		executeAction(stringIDToTypeID("placeEvent"), desc, DialogModes.ERROR);
	}

	function moveTo(x, y) {
		var desc2 = new ActionDescriptor();
		desc2.putUnitDouble(stringIDToTypeID("horizontal"), stringIDToTypeID("pixelsUnit"), x);
		desc2.putUnitDouble(stringIDToTypeID("vertical"), stringIDToTypeID("pixelsUnit"), y);

		var desc = new ActionDescriptor();
		desc.putObject(stringIDToTypeID("position"), stringIDToTypeID("point"), desc2);
		desc.putBoolean(stringIDToTypeID("relative"), false);
		executeAction(stringIDToTypeID("transform"), desc, DialogModes.ERROR);
	}


	try {
		var startRulerUnits = app.preferences.rulerUnits;
		app.preferences.rulerUnits = Units.PIXELS;

		place(filename);
		var imageWidth = Number(activeDocument.activeLayer.bounds[2]) - Number(activeDocument.activeLayer.bounds[0]);
		var imageHeight = Number(activeDocument.activeLayer.bounds[3]) - Number(activeDocument.activeLayer.bounds[1]);

		activeDocument.activeLayer.resize(newWidth / imageWidth * 100, newHeight / imageHeight * 100);

		moveTo(x - newWidth / 2 - 60, y - newHeight / 2 - 60);
		activeDocument.activeLayer.name = layerName;

		app.preferences.rulerUnits = startRulerUnits;
	} catch (e) {
		alert(e);
	}
}

pasteImage(arguments[0], arguments[1], arguments[2], arguments[3], arguments[4], arguments[5])