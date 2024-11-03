// FROM clause - specifies the data source path
from: "service.frontend"

// SELECT clause - fields to project
select: [
    "*"    // Select all fields from the matched path
]

// WHERE clause - predicate conditions
where: {
    dependencies: ["database"] 
}