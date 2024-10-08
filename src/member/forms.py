from django import forms
from django.contrib.auth import get_user_model
from django.core.exceptions import ValidationError
from django.utils.translation import gettext_lazy as _

from .models import Profile


class AddMemberForm(forms.Form):
    username = forms.CharField()
    password = forms.CharField(
        label=_("Password"), strip=False, widget=forms.PasswordInput
    )
    user_model = get_user_model()

    def clean_username(self):
        username = self.cleaned_data.get("username")
        if (
            username
            and self.user_model.objects.filter(username__iexact=username).exists()
        ):
            self.add_error(
                "username", ValidationError(_("The username is not unique"), "username")
            )
        else:
            return username

    def save(self):
        user = self.user_model.objects.create_user(
            self.cleaned_data.get("username"),
            password=self.cleaned_data.get("password"),
        )
        user.save()
        profile = Profile(user=user)
        profile.save()
        return user


class UpdateMemberForm(forms.Form):
    about = forms.CharField(widget=forms.Textarea, required=False)
    email = forms.EmailField(required=False)
    user_model = get_user_model()

    def __init__(self, member, *args, **kwargs):
        self.member = member
        super().__init__(*args, **kwargs)

    def clean_email(self):
        email = self.cleaned_data.get("email")

        if (
            email
            and "email" in self.changed_data
            and self.user_model.objects.filter(email__iexact=email).exists()
        ):
            self.add_error(
                "email", ValidationError(_("The email is not unique"), "email")
            )
        else:
            return email

    def save(self):
        user = self.member
        if not self.has_changed():
            return user

        user.email = self.cleaned_data.get("email")
        user.profile.about = self.cleaned_data.get("about")
        user.profile.save()
        user.save()
        return self.member
