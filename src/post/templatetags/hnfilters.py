from django import template


register = template.Library()

@register.filter(name="indent")
def indent(value):
    return f"text-indent: {value}em"
