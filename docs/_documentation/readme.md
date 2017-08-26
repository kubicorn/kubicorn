---
layout: documentation
title: Kubicorn Documentation
date: 2017-08-25
doctype: general
---

## Rules

### Know what you are writing

You either are writing cloud specific docs or you aren't.

If you are writing documentation for a cloud, please treat it as such.

Cloud docs have a `doctype: [cloud]` property, and talk about specific clouds. The property can be `aws`, `azure`, `do`, or `google`.

If you are writing docs for global concepts, never mention a cloud. In this case you should use `doctype: general`.

All documentation files should include the following header:

```
---
layout: documentation
title: [The title of your document]
date: YYYY-MM-DD
doctype: [general/aws/azure/do/google]
---
```

### Keep tables formatted

If you are making a table in markdown: YES, we expect you to format them nicely. Keep our shit clean please.

### Complete sentences

A complete sentence expresses a complete thought.

Always write complete sentences.

Always use proper grammar.

Each sentence goes on a new line.

### Write inclusively

Always use "we" when referring to the project, and always refer to the user as "the user".


### Must and Might

Use the word MUST to explain something a user has to accomplish in your documentation.

Use the word MIGHT to explain something a user optionally can chose to accomplish in your documentation.

Never use the word MAY.
