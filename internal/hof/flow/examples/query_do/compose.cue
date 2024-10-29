common:{
    output: "outfile.txt"
},
commands: {
    step1: ["run","--apply","./query_resources.tf.cue"]
    step2: [ "codegen","--code-dir", "./output","--prompt", "create an aws resource similar to this resource","--context","\(common.output)","--system-prompt","terraform_prompt.md"]
    step3: ["validate", "--code-dir","./output"]
}