// FROM clause - specifies the data source path
from: "service.web"

// SELECT clause - fields to project
select: [
    "name"    // Select all fields from the matched path
]

// WHERE clause - predicate conditions
where: {
    // name: "web-*" 
}
