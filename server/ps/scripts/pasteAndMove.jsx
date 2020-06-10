function pasteImage(filename, layerName, x, y) {

	function place(filename) {
		var desc = new ActionDescriptor();	
		desc.putPath(stringIDToTypeID("null"), new File( filename));
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
		place(filename);
		moveTo(x, y);		
		activeDocument.activeLayer.name = layerName; 
		
	} catch (e) {
		alert(e);
	}
}

pasteImage(arguments[0], arguments[1], arguments[2], arguments[3])