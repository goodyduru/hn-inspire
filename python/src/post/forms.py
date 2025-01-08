from django import forms
from django.core.exceptions import ValidationError
from django.utils.translation import gettext_lazy as _

from .models import Post


class AddPost(forms.Form):
    title = forms.CharField(label=_("Title"))
    url = forms.URLField(required=False, label=_("URL"))
    text = forms.CharField(widget=forms.Textarea, required=False, label=_("Text"))

    def __init__(self, author, *args, **kwargs):
        self.author = author
        super().__init__(*args, **kwargs)

    def clean(self):
        url = self.cleaned_data.get("url")
        text = self.cleaned_data.get("text")
        if url == "" and text == "":
            raise ValidationError(
                _("URL and Text cannot be empty at the same time"), "post_error"
            )
        return self.cleaned_data

    def save(self):
        url = self.cleaned_data.get("url")
        text = self.cleaned_data.get("text")
        title = self.cleaned_data.get("title")
        post = Post(title=title, url=url, text=text, author=self.author, votes=1)
        post.save()
        return post


class AddComment(forms.Form):
    text = forms.CharField(
        widget=forms.Textarea(attrs={"cols": "80", "rows": "8"}), label=_("")
    )
    parent = forms.CharField(widget=forms.HiddenInput)

    def __init__(self, author, *args, **kwargs):
        self.author = author
        super().__init__(*args, **kwargs)

    def clean(self):
        parent = self.cleaned_data.get("parent")
        if not parent.isdigit():
            raise ValidationError(_("Parent cannot be empty"), "comment_error")
        try:
            self.post_parent = Post.objects.get(pk=int(parent))
        except:
            raise ValidationError(_("Parent cannot be invalid"), "comment_error")

        return self.cleaned_data

    def save(self):
        text = self.cleaned_data.get("text")
        post = Post(text=text, author=self.author, parent=self.post_parent)
        post.save()
        return post
