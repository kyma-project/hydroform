package workspace

const handlerPython = `def main(event, context):
    return "hello world"`

const (
	FileNameHandlerPy       FileName = "handler.py"
	FileNameRequirementsTxt FileName = "requirements.txt"
)

var workspacePython = workspace{
	newTemplatedFile(handlerPython, FileNameHandlerPy),
	newEmptyFile(FileNameRequirementsTxt),
}
