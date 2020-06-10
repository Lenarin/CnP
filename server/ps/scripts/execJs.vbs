Dim appRef
Set appRef = CreateObject("Photoshop.Application")
if wScript.Arguments.Count = 0 then
    wScript.Echo "No params"
else
    path = wScript.Arguments(0)
    args = wScript.Arguments(1)
    error = appRef.DoJavaScriptFile(path, Split(args, "|"))
    if Not error = "true" and Not error = "[ActionDescriptor]" and Not error = "undefined" then
        Err.raise 1, "execJs.vbc", error
    end if
end if