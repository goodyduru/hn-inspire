from django.contrib.auth import get_user_model
from django.db import models


# Create your models here.
class Post(models.Model):
    title = models.CharField(max_length=255, default="", blank=True)
    url = models.URLField(max_length=255, default="", blank=True)
    text = models.TextField(blank=True, default="")
    author = models.ForeignKey(get_user_model(), on_delete=models.SET_NULL, null=True)
    parent = models.ForeignKey("self", on_delete=models.SET_NULL, null=True, blank=True)
    votes = models.IntegerField(default=0)
    is_flagged = models.BooleanField(default=False)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    def is_comment(self):
        return self.parent is not None

    def __str__(self) -> str:
        return f"{self.title}"
