function pasteImage(filename, layerName, x, y) {

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

	function getBounds() {
		var ref = new ActionReference();
		ref.putProperty(stringIDToTypeID("property"), stringIDToTypeID("viewInfo"));
		ref.putEnumerated(stringIDToTypeID("document"), stringIDToTypeID("ordinal"), stringIDToTypeID("targetEnum"));
		var desc = executeActionGet(ref);

		var bounds = desc.getObjectValue(stringIDToTypeID('viewInfo')).getObjectValue(stringIDToTypeID('activeView')).getObjectValue(stringIDToTypeID('globalBounds'));

		var left = bounds.getDouble(stringIDToTypeID('left'));
		var right = bounds.getDouble(stringIDToTypeID('right'))
		var top = bounds.getDouble(stringIDToTypeID('top'))
		var bottom = bounds.getDouble(stringIDToTypeID('bottom'))

		return { left: left, right: right, top: top, bottom: bottom }
	}


	try {
		var startRulerUnits = app.preferences.rulerUnits;
		app.preferences.rulerUnits = Units.PIXELS;

		var bounds = getBounds()

		alert(bounds.left + '\n' + bounds.right + '\n' + bounds.top + '\n' + bounds.bottom + '\n');

		place(filename);
		var imageWidhth = Number(activeDocument.activeLayer.bounds[2]) - Number(activeDocument.activeLayer.bounds[0]);
		var imageHeight = Number(activeDocument.activeLayer.bounds[3]) - Number(activeDocument.activeLayer.bounds[1]);

		moveTo(x - imageWidhth / 2, y - imageHeight / 2);
		activeDocument.activeLayer.name = layerName;
		//activeDocument.activeLayer.resizeImage(50, 50);
		//activeDocument.activeLayer.resizeCanvas(50, 50);

		app.preferences.rulerUnits = startRulerUnits;
	} catch (e) {
		alert(e);
	}
}

pasteImage(arguments[0], arguments[1], arguments[2], arguments[3])