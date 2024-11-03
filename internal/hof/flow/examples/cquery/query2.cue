from : "service[string]"
select: [
    "name",
    "_file",
    "dependencies"
]
where: {
    // To match services with 'database' as a dependency, you could use:
    // "name": ["frontend"]  // Using the =~ operator for regex matching
    dependencies: ["cache"]
}