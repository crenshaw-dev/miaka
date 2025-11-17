```yaml
apiVersion: example.com/v1alpha1
kind: Example
# +required
spec:
  count: 1  # This type will be inferred to be int
  message: "This will be inferred to be a string"
  # This comment will be applied to the list field.
  list:
  # This comment will be applied to the struct.
    # This comment will be applied to the first field in the object since it's indented to match.
    # It's identical to the comment in the second list item. Otherwise an error would be thrown.
  - option1: value
    # option2 is another option.
    option2: value
    # This comment will be applied to the first field in the object since it's indented to match.
    # It's identical to the comment in the second list item. Otherwise an error would be thrown.
  - option1: value
    # option3 is a third option. This field will be added to the list item struct with the other 2.
  - option3: value
```