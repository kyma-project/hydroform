package workspace

const handlerPython = `def main(event, context):
    return "hello world"`

const (
	FileNameHandlerPy       FileName = "handler.py"
	FileNameRequirementsTxt FileName = "requirements.txt"
)

var workspacePython = workspace{
	NewTemplatedFile(handlerPython, FileNameHandlerPy),
	newEmptyFile(FileNameRequirementsTxt),
}
